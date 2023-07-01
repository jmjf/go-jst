package common

import (
	"strings"
	"time"
)

type Date time.Time

// NewDate attempts to convert a string to a Date (date only part of a time.Time)
// Returns a Date and nil error on success or indeterminate Date and error on failure.
func NewDate(str string) (Date, error) {
	dt, err := time.Parse(time.DateOnly, str)
	if err != nil {
		return Date(dt), err
	}
	return Date(dt), nil
}

// NewDateFromTime converts a time.Time to a Date (date only part of a time.Time)
// Returns a Date (time.Time can always be converted successfully).
func NewDateFromTime(tm time.Time) Date {
	dt, _ := time.Parse(time.DateOnly, tm.Format(time.DateOnly))
	return Date(dt)
}

// Date.String() formats a Date as a date only string and returns the string.
//
// Mutates receiver: no
func (dt Date) String() string {
	return time.Time(dt).Format(time.DateOnly)
}

// Date.UnmarshalJSON attempts to convert a JSON time to a Date.
// On success, it sets Date to the date only part of the time and returns a nil error.
// On failure, it returns an error from time.Parse().
//
// Mutates receiver: yes
func (dt *Date) UnmarshalJSON(b []byte) error {
	// remove leading/trailing "
	str := strings.Trim(string(b), "\"")
	t, err := time.Parse(time.DateOnly, str)
	if err != nil {
		return err
	}
	*dt = Date(t)
	return nil
}

// Date.MarshalJSON converts a Date to a string to a quote-wrapped byte slice.
// On success, it returns a byte slice and a nil error.
// It cannot fail (Date can always be converted to a string).
//
// Mutates receiver: no
func (dt Date) MarshalJSON() ([]byte, error) {
	str := dt.String()
	return []byte("\"" + str + "\""), nil
}
