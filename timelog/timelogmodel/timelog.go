package timelogmodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/rebel-l/go-utils/slice"
)

const (
	// LocationAbsence defines a pseudo-location to show absences. It means timelog doesn't belong anywhere.
	LocationAbsence = "absence"

	// LocationHome defines the possible value for the homeoffice location. It means timelog happened at home.
	LocationHome = "home"

	// LocationOffice defines the possible value for the office location. It means timelog happened at the office.
	LocationOffice = "office"

	// ReasonBreak defines the value for the timelog reason having a break.
	ReasonBreak = "break"

	// ReasonWork defines the value for the timelog reason that this entry is work time.
	ReasonWork = "work"

	// ReasonSickLeave defines the value for the timelog reason that this entry is sick leave.
	ReasonSickLeave = "sick leave"

	// ReasonVacation defines the value for the timelog reason that this entry is vacation.
	ReasonVacation = "vacation"
)

var (
	// ErrDecodeJSON occurs if a string is not in JSON format.
	ErrDecodeJSON = errors.New("failed to decode JSON")

	// ErrValidationStartMandatory occurs during validation if the Start time wasn't set.
	ErrValidationStartMandatory = fmt.Errorf("start time should not be empty")

	// ErrValidationInvalidLocation occurs during validation if the location is not one of the known ones.
	ErrValidationInvalidLocation = fmt.Errorf("location must be one of the following values")

	// ErrValidationInvalidReason occurs during validation if the reason is not one of the known ones.
	ErrValidationInvalidReason = fmt.Errorf("reason must be one of the following values")

	locations = slice.StringSlice{
		LocationAbsence,
		LocationHome,
		LocationOffice,
	}

	reasons = slice.StringSlice{
		ReasonWork,
		ReasonBreak,
		ReasonVacation,
		ReasonSickLeave,
	}
)

// Timelog represents a model of repository including business logic.
type Timelog struct {
	ID         uuid.UUID  `json:"ID"`
	Start      time.Time  `json:"Start"`
	Stop       *time.Time `json:"Stop,omitempty"`
	Reason     string     `json:"Reason"`
	Location   string     `json:"Location"`
	CreatedAt  time.Time  `json:"CreatedAt"`
	ModifiedAt time.Time  `json:"ModifiedAt"`
}

// DecodeJSON converts JSON data to struct.
func (t *Timelog) DecodeJSON(reader io.Reader) error {
	if t == nil {
		return nil
	}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(t); err != nil {
		return fmt.Errorf("%w: %v", ErrDecodeJSON, err)
	}

	return nil
}

// Validate is validating the attributes of the struct to valid values. If the validation fails it returns the reason
// why it failed in the error message.
func (t *Timelog) Validate() error {
	if t.Start.IsZero() {
		return ErrValidationStartMandatory
	}

	if locations.IsNotIn(t.Location) {
		return fmt.Errorf("%w: %s", ErrValidationInvalidLocation, locations.String())
	}

	if reasons.IsNotIn(t.Reason) {
		return fmt.Errorf("%w: %s", ErrValidationInvalidReason, reasons.String())
	}

	return nil
}
