package bettererror

import (
	"encoding/json"
	"log"
	"time"
)

type IBetterError interface {
	JSON() []byte
	SetTime() IBetterError
	SetAppLayer(layer string) IBetterError
}

type BetterError struct {
	Measure *time.Time `json:"measure"`
	Layer   string     `json:"layer,omitempty"`
	Err     string     `json:"err"`
}

const Storage string = "storage"
const Logic string = "business logic"
const Handler string = "handler"

// New allows to modify error that would implement IBetterError
func New(err error) IBetterError {
	if err == nil {
		return nil
	}
	t := time.Now()
	return &BetterError{Err: err.Error(), Measure: &t}
}

func (e *BetterError) SetTime() IBetterError {
	Time := time.Now()
	e.Measure = &Time
	return e
}

func (e *BetterError) Time() time.Time {
	return *e.Measure
}

func (e *BetterError) SetAppLayer(layer string) IBetterError {
	e.Layer = layer
	return e
}

func (e *BetterError) JSON() []byte {
	marshal, err := json.Marshal(e)
	if err != nil {
		log.Println("json.Marshal:", err)
	}

	return marshal
}
