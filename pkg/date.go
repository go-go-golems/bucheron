package pkg

import (
	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	"github.com/tj/go-naturaldate"
	"time"
)

// refTime is used to set a reference time for natural date parsing for unit test purposes
var refTime *time.Time

func ParseDate(value string) (time.Time, error) {
	parsedDate, err := dateparse.ParseAny(value)
	if err != nil {
		refTime_ := time.Now()
		if refTime != nil {
			refTime_ = *refTime
		}
		parsedDate, err = naturaldate.Parse(value, refTime_)
		if err != nil {
			return time.Time{}, errors.Wrapf(err, "Could not parse date: %s", value)
		}
	}

	return parsedDate, nil
}
