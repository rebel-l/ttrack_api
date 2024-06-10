package publicholidaymodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

// ErrDecodeJSON occurs if a string is not in JSON format.
var ErrDecodeJSON = errors.New("failed to decode JSON")

// PublicHoliday represents a model of repository including business logic.
type PublicHoliday struct {
	ID         uuid.UUID `json:"ID"`
	Day        time.Time `json:"Day"`
	Name       string    `json:"Name"`
	HalfDay    bool      `json:"HalfDay"`
	CreatedAt  time.Time `json:"CreatedAt"`
	ModifiedAt time.Time `json:"ModifiedAt"`
}

// DecodeJSON converts JSON data to struct.
func (r *PublicHoliday) DecodeJSON(reader io.Reader) error {
	if r == nil {
		return nil
	}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(r); err != nil {
		return fmt.Errorf("%w: %v", ErrDecodeJSON, err)
	}

	return nil
}
