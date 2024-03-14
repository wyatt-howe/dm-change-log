package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"

	"github.com/carbondmp/initialize"
)

const (
	serviceName = "dm-change-log/consumer"
)

var (
	client      = &http.Client{}
	apiRoot     string
	version     = "dev"
	buildCommit = ""
	buildDate   = ""
	buildBy     = ""
)

func main() {
	log.Printf("Starting %s. Version: %s. Build Commit: %s. Build Date: %s. Build By: %s.\n",
		serviceName, version, buildCommit, buildDate, buildBy)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: viper.GetStringSlice("kafka_brokers"),
		Topic:   viper.GetString("kafka_topic"),
		GroupID: viper.GetString("kafka_group_id"),
		MaxWait: 10 * time.Millisecond,
	})

	log.Printf("change-log-api-worker ready")
	ctx := context.TODO()
	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println(string(msg.Value))

		func() {
			resp, err := client.Post(apiRoot+"/change-events", "application/json", bytes.NewReader(msg.Value))
			if err != nil {
				log.Println(err)
				return
			}
			defer resp.Body.Close()

			log.Println(resp.Status)
		}()
	}
}

func init() {
	err := initialize.Service()
	fatal(err)

	apiRoot = viper.GetString("api_root")

	resp, err := client.Get(apiRoot + "/accessible")
	fatal(err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("status code != 200 (%s)", resp.Status)
	}

	var ok bool
	err = json.NewDecoder(resp.Body).Decode(&ok)
	fatal(err)

	if !ok {
		log.Fatalf("was able to reach api but it responded !ok")
	}
}

func fatal(err error) {
	if err != nil {
		log.SetFlags(0)
		_, f, l, _ := runtime.Caller(1)
		log.Fatalf("%s:%d %s", f, l, err)
	}
}
