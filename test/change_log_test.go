package test_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/segmentio/kafka-go"

	"github.com/carbondmp/dm-change-log/model"
)

func TestIntegration(t *testing.T) {
	fatal := fatalFactory(t)

	testChangeEvent := model.ChangeEvent{
		ID:                 ulid.Make(),
		EventTime:          time.Now().Unix(),
		EventObjectID:   ulid.Make().String(),
		EventObjectType: "soul",
		EffectedService:    "valhalla",
		SourceService:      "battle",
		User:               "thor",
		Reason:             "bordem",
		EventType:          "create",
	}

	{
		w := &kafka.Writer{
			Addr:         kafka.TCP("localhost:9092"),
			Topic:        "dm-change-log",
			BatchTimeout: time.Millisecond,
		}

		body, err := json.MarshalIndent(testChangeEvent, "", "\t")
		fatal(err)

		err = w.WriteMessages(context.TODO(), kafka.Message{
			Key:   testChangeEvent.ID[:],
			Value: body,
		})
		fatal(err)

		t.Log("wrote message to kafka")
	}

	log.Println("sleeping to allow worker to process the record")
	time.Sleep(time.Second)

	{
		resp, err := http.Get(fmt.Sprintf("http://localhost:12345/change-log/v1/change-event/%s", testChangeEvent.ID))
		fatal(err)
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Fatalf("received != 200 response status '%s'", resp.Status)
		}

		var changeEvent model.ChangeEvent
		err = json.NewDecoder(resp.Body).Decode(&changeEvent)
		fatal(err)

		log.Println(changeEvent)
		if !reflect.DeepEqual(testChangeEvent, changeEvent) {
			t.Fatal("failed deep equality")
		}
	}
}

func fatalFactory(t *testing.T) func(error) {
	return func(err error) {
		if err != nil {
			_, f, l, _ := runtime.Caller(1)
			t.Fatalf("%s:%d %s", f, l, err)
		}
	}
}
