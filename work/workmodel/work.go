package workmodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

// ErrDecodeJSON occurs if the a string is not in JSON format.
var ErrDecodeJSON = errors.New("failed to decode JSON")

// Work represents a model of repository including business logic.
type Work struct {
	ID         uuid.UUID `json:"ID"`
	Start      time.Time `json:"Start"`
	Stop       time.Time `json:"Stop"`
	CreatedAt  time.Time `json:"CreatedAt"`
	ModifiedAt time.Time `json:"ModifiedAt"`
}

// DecodeJSON converts JSON data to struct.
func (r *Work) DecodeJSON(reader io.Reader) error {
	if r == nil {
		return nil
	}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(r); err != nil {
		return fmt.Errorf("%w: %v", ErrDecodeJSON, err)
	}

	return nil
}
