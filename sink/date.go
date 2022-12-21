// Copyright © 2022 Sloan Childers
package sink

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gertd/go-pluralize"
	"github.com/rs/zerolog/log"
)

var ErrEmptyDate = errors.New("empty date string")
var ErrParseDate = errors.New("non-parsable date string")

type DateStuff struct {
	pluralize *pluralize.Client
}

func NewDateStuff() *DateStuff {
	return &DateStuff{pluralize: pluralize.NewClient()}
}

// FORMAT:  YYYY-MM-DD
func GetDaysFromDate(date string) (int, error) {
	if date == "" {
		return 0, ErrEmptyDate
	}

	if strings.Contains(date, "T") {
		parts := strings.Split(date, "T")
		if len(parts) > 1 {
			date = parts[0]
		}
	} else if strings.Contains(date, " ") {
		parts := strings.Split(date, " ")
		if len(parts) > 1 {
			date = parts[0]
		}
	}

	dt, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Error().Err(err).Str("component", "date").Str("date", date).Msg("parse date")
		return 0, ErrParseDate
	}
	diff := time.Since(dt)
	return int(diff.Hours() / 24), nil
}

// 4.days.ago, 1.year.ago, etc. to a date string
func (x *DateStuff) AgoStringToDate(date string) string {
	parts := strings.Split(date, ".")
	duration, _ := strconv.Atoi(parts[0])
	sign := 1
	if parts[2] == strings.ToLower("ago") {
		sign = -1
	}
	days, months, years := 0, 0, 0
	switch x.pluralize.Plural(parts[1]) {
	case "days":
		days = duration * sign
	case "months":
		months = duration * sign
	case "years":
		years = duration * sign
	}
	before := time.Now().AddDate(years, months, days)
	return before.Format("2006-01-02")
}
