package common

import (
	"strings"
	"time"
)

type Date time.Time

func NewDate(str string) (Date, error) {
	dt, err := time.Parse(time.DateOnly, str)
	if err != nil {
		return Date(dt), err
	}
	return Date(dt), nil
}

func NewDateFromTime(tm time.Time) Date {
	dt, _ := time.Parse(time.DateOnly, tm.Format(time.DateOnly))
	return Date(dt)
}

func (dt Date) String() string {
	return time.Time(dt).Format(time.DateOnly)
}

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

func (dt Date) MarshalJSON() ([]byte, error) {
	str := dt.String()
	return []byte("\"" + str + "\""), nil
}
