package main

import (
	"encoding/json"
	"log"
	"runtime"
	"time"

	"github.com/carbondmp/initialize"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"

	"github.com/carbondmp/dm-change-log/model"
)

const (
	serviceName = "dm-change-log/emitter"
)

var (
	w           *kafka.Writer
	version     = "dev"
	buildCommit = ""
	buildDate   = ""
	buildBy     = ""
)

func init() {
	err := initialize.Service()
	fatal(err)

	w = &kafka.Writer{
		Addr:         kafka.TCP(viper.GetStringSlice("kafka_brokers")...),
		Topic:        viper.GetString("kafka_topic"),
		BatchTimeout: 10 * time.Millisecond,
	}

	gin.SetMode(gin.ReleaseMode)
}

func main() {
	log.Printf("Starting %s. Version: %s. Build Commit: %s. Build Date: %s. Build By: %s.\n",
		serviceName, version, buildCommit, buildDate, buildBy)
	r := gin.New()
	r.UseH2C = true
	r.Use(gin.Recovery())
	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": version, "build_commit": buildCommit, "build_date": buildDate, "build_by": buildBy})
	})
	r.Use(gin.Logger())

	v1 := r.Group("/change-log/v1/emitter")
	v1.GET("/accessible", func(c *gin.Context) { c.JSON(200, true) })
	v1.POST("/change-events", createChangeEvents)

	port := ":" + viper.GetString("port")
	log.Println("dm-change-log-emitter listening on", port)
	log.Fatal(r.Run(port))
}

func createChangeEvents(c *gin.Context) {
	var event model.ChangeEvent
	err := c.Bind(&event)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(400, gin.H{"message": "failed to deserialize into change event", "error": err.Error()})
		return
	}

	body, err := json.Marshal(event)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(428, gin.H{"message": "failed to serialize change event to json", "error": err.Error()})
		return
	}

	err = w.WriteMessages(c.Request.Context(), kafka.Message{Key: event.ID[:], Value: body})
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(500, gin.H{"message": "failed to write change event to kafka", "error": err.Error()})
		return
	}

	c.JSON(202, event)
}

func fatal(err error) {
	if err != nil {
		log.SetFlags(0)
		_, f, l, _ := runtime.Caller(1)
		log.Fatalf("%s:%d %s", f, l, err)
	}
}
