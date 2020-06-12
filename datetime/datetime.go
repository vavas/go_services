// Package datetime is a library for date & time manipulation.
package datetime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// DT represents time.Time.
// When marshalled to JSON, it will return unix timestamp.
// When unmarshalled from JSON, it can parse various date time format.
type DT struct {
	time.Time
}

// MarshalJSON implements the json.Marshaler interface.
func (dt DT) MarshalJSON() (result []byte, err error) {
	unix := dt.UnixNano()
	if unix > 0 {
		return []byte(fmt.Sprintf("%.9f", float64(unix)/float64(1e9))), nil
	}
	return []byte("0"), nil
}

// MarshalBSONValue implements the bson.ValueMarshaler interface.
func (dt DT) MarshalBSONValue() (resultType bsontype.Type, resultBytes []byte, err error) {
	return bson.MarshalValue(dt.Time)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (dt *DT) UnmarshalJSON(data []byte) error {
	tPtr, err := ParseTimeString(string(data))
	if err != nil {
		return err
	}

	*dt = DT{*tPtr}
	return nil
}

//UnmarshalBSONValue implements the bson.ValueUnmarshaler interface.
func (dt *DT) UnmarshalBSONValue(tp bsontype.Type, b []byte) (err error) {
	rawValue := bson.RawValue{Type: tp, Value: b}

	tm, ok := rawValue.TimeOK()
	if !ok {
		return errors.New("can not decode into a Time")
	}

	*dt = DT{tm}

	return nil
}

// ParseTimeString parses a string to a time.Time object.
func ParseTimeString(strData string) (*time.Time, error) {
	if strings.HasPrefix(strData, `"`) {
		strData = strData[1:]
	}
	if strings.HasSuffix(strData, `"`) {
		strData = strData[:len(strData)-1]
	}

	f, err := strconv.ParseFloat(strData, 64)
	if err == nil {
		sec := int64(f)
		nsec := int64((f - float64(sec)) * 1e+9)
		t := time.Unix(sec, nsec)
		return &t, nil
	}

	t, err := dateparse.ParseAny(strData)
	if err == nil {
		return &t, nil
	}

	timeFormats := []string{
		// ISO 8601ish formats
		time.RFC3339Nano,
		time.RFC3339,

		// Unusual formats, prefer formats with timezones
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.UnixDate,
		time.RubyDate,
		time.ANSIC,

		// Hilariously, Go doesn't have a const for it's own time layout.
		// See: https://code.google.com/p/go/issues/detail?id=6587
		"2006-01-02 15:04:05.999999999 -0700 MST",

		"2006-01-02T15:04:05.999999999", // RFC3339Nano without timezone
		"2006-01-02T15:04:05",           // RFC3339 without timezone

		"2006-01-02 15:04:05.999999999 -07:00",
		"2006-01-02 15:04:05.999999999 -0700",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05 -07:00",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05",

		"Mon Jan 2 2006 15:04:05 GMT-0700 (MST)", // Javascript Date.toString()
	}

	for _, format := range timeFormats {
		t, err := time.Parse(format, strData)
		if err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf(`Cannot parse "%s"`, strData)
}
