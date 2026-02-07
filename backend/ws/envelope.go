package ws

import (
	"encoding/json"
	"log"
)

type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func Marshal(data interface{}) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("Marshal failed: %v", err)
		return []byte("{}")
	}
	return b
}

func EnvelopeMarshal(msgType string, data interface{}) []byte {
	envelope := map[string]interface{}{
		"type": msgType,
		"data": data,
	}
	return Marshal(envelope)
}

func Unmarshal(data []byte, v any) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		log.Printf("Unmarshal failed: %v", err)
		return err
	}
	return nil
}
