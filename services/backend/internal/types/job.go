package types

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Job struct {
	ID         uuid.UUID
	Kind       string
	Queue      string
	Payload    json.RawMessage
	State      string
	Priority   int
	Attempt    int
	MaxAttempt int
}
