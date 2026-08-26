package serialization

import (
	"encoding/json"
	"errors"
	"time"
)

type Envelope struct {
	Event     string          `json:"event"`
	RequestID string          `json:"request_id"`
	At        time.Time       `json:"at"`
	Payload   json.RawMessage `json:"payload"`
}

func Encode(event, request string, payload any, at time.Time) ([]byte, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Event: event, RequestID: request, At: at.UTC(), Payload: b})
}
func Decode(data []byte) (Envelope, error) {
	var out Envelope
	if err := json.Unmarshal(data, &out); err != nil {
		return out, err
	}
	if out.Event == "" || out.RequestID == "" || len(out.Payload) == 0 {
		return out, errors.New("invalid envelope")
	}
	return out, nil
}
func DecodePayload(e Envelope, target any) error      { return json.Unmarshal(e.Payload, target) }
func ClonePayload(in json.RawMessage) json.RawMessage { return append(json.RawMessage(nil), in...) }
