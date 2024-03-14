package model

import (
	"github.com/oklog/ulid/v2"
)

type EventType string

const (
	EventTypeCreate      = "create"
	EventTypeUpdate      = "update"
	EventTypeSoftDelete  = "soft_delete"
	EventTypeSoftRestore = "soft_restore"
	EventTypeHardDelete  = "hard_delete"
)

type ChangeEvent struct {
	ID                 ulid.ULID `json:"id"`
	EventTime          int64     `json:"event_time"`
	EventObjectID   string    `json:"event_object_id"`
	EventObjectType string    `json:"event_object_type"`
	EffectedService    string    `json:"effected_service"`
	SourceService      string    `json:"source_service"`
	CorrelationID      string    `json:"correlation_id"`
	User               string    `json:"user"`
	Reason             string    `json:"reason"`
	Comment            string    `json:"comment"`
	EventType          `json:"event_type"`
	BeforeObject       interface{} `json:"before_object"`
	AfterObject        interface{} `json:"after_object"`
}
