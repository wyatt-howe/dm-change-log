package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"runtime"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/viper"

	"github.com/carbondmp/initialize"
	"github.com/wyatt-howe/dm-change-log/model"
)

const (
	serviceName = "dm-change-log/api"
)

var (
	db          *DB
	version     = "dev"
	buildCommit = ""
	buildDate   = ""
	buildBy     = ""
)

func init() {
	err := initialize.Service()
	fatal(err)

	db, err = NewDB(viper.GetString("db_main"), viper.GetString("db_read_only"), viper.GetInt("db_max_conn"))
	fatal(err)

	gin.SetMode(gin.ReleaseMode)
}

func main() {
	log.Printf("Starting %s. Version: %s. Build Commit: %s. Build Date: %s. Build By: %s.\n",
		serviceName, version, buildCommit, buildDate, buildBy)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": version, "build_commit": buildCommit, "build_date": buildDate, "build_by": buildBy})
	})

	r.Use(gin.Logger())

	v1 := r.Group("/change-log/v1")
	v1.GET("/accessible", func(c *gin.Context) { c.JSON(200, true) })
	v1.GET("/change-events", getChangeEvents)
	v1.GET("/change-event/:id", getChangeEventByID)
	v1.POST("/change-events", createChangeEvents)

	port := ":" + viper.GetString("port")
	log.Println("dm-change-log listening on", port)
	log.Fatal(r.Run(port))
}

func getChangeEvents(c *gin.Context) {

	rows, err := db.ReadOnly.QueryContext(c.Request.Context(), "SELECT id, event_time, event_object_id, event_object_type, effected_service, source_service, correlation_id, user, reason, comment, event_type, before_object, after_object FROM change_event")
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(500, gin.H{"message": "failed to hit change_event table", "error": err.Error()})
		return
	}
	defer rows.Close()

	res := []model.ChangeEvent{}

	for rows.Next() {
		var event model.ChangeEvent
		err = rows.Scan(&event.ID, &event.EventTime, &event.EventObjectID, &event.EventObjectType, &event.EffectedService, &event.SourceService, &event.CorrelationID, &event.User, &event.Reason, &event.Comment, &event.EventType, &event.BeforeObject, &event.AfterObject)
		if err != nil {
			log.Println(err)
			continue
		}
		res = append(res, event)
	}

	c.JSON(200, res)
}

func getChangeEventByID(c *gin.Context) {
	reqID, err := ulid.Parse(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(400, gin.H{"message": "id was not a valid ulid", "error": err.Error()})
		return
	}

	var event model.ChangeEvent
	err = db.ReadOnly.QueryRowContext(c.Request.Context(), "SELECT id, event_time, event_object_id, event_object_type, effected_service, source_service, correlation_id, user, reason, comment, event_type, before_object, after_object FROM change_event WHERE id = ?", reqID[:]).Scan(&event.ID, &event.EventTime, &event.EventObjectID, &event.EventObjectType, &event.EffectedService, &event.SourceService, &event.CorrelationID, &event.User, &event.Reason, &event.Comment, &event.EventType, &event.BeforeObject, &event.AfterObject)
	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatusJSON(404, gin.H{"message": fmt.Sprintf("change_event with id '%s' not found", reqID)})
		} else {
			log.Println(err)
			c.AbortWithStatusJSON(500, gin.H{"message": "failed to hit change_event table", "error": err.Error()})
		}
		return
	}

	c.JSON(200, event)
}

func createChangeEvents(c *gin.Context) {
	var event model.ChangeEvent
	err := c.Bind(&event)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(400, gin.H{"message": "failed to deserialize into change event", "error": err.Error()})
		return
	}

	if event.ID.Compare(ulid.ULID{}) == 0 {
		event.ID = ulid.Make()
	}

	eventBefore, err := json.Marshal(event.BeforeObject)
	fatal(err)

	eventAfter, err := json.Marshal(event.AfterObject)
	fatal(err)

	_, err = db.ExecContext(c.Request.Context(),
		"INSERT INTO change_event VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		event.ID[:], event.EventTime, event.EventObjectID, event.EventObjectType, event.EffectedService, event.SourceService, event.CorrelationID, event.User, event.Reason, event.Comment, event.EventType, eventBefore, eventAfter)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(500, gin.H{"message": "failed to hit change_event table", "error": err.Error()})
		return
	}

	c.JSON(201, event)
}

func fatal(err error) {
	if err != nil {
		log.SetFlags(0)
		_, f, l, _ := runtime.Caller(1)
		log.Fatalf("%s:%d %s", f, l, err)
	}
}
