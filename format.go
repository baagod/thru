package warp

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	Layout      = "01/02 03:04:05PM '06 -0700" // The reference time, in numerical order.
	ANSIC       = "Mon Jan _2 15:04:05 2006"
	UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
	RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
	RFC822      = "02 Jan 06 15:04 MST"
	RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
	RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
	RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	RFC3339     = "2006-01-02T15:04:05Z07:00"
	RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
	Kitchen     = "3:04PM"
	Stamp       = "Jan _2 15:04:05"
	StampMilli  = "Jan _2 15:04:05.000"
	StampMicro  = "Jan _2 15:04:05.000000"
	StampNano   = "Jan _2 15:04:05.000000000"
	DateTime    = "2006-01-02 15:04:05"
	DateOnly    = "2006-01-02"
	TimeOnly    = "15:04:05"
)

const (
	dateonly = `\d{4}(-\d{2}){2}`
	datetime = `(\d{2}:){2}\d{2}(\.\d{1,9})?`
	mst      = `[A-Z]{3,4}([+\-]\d{1,2})?`
	z0700    = `[+\-]\d{4}`
)

var patterns = map[string]*regexp.Regexp{
	"2006-01-02T15:04:05.999999999Z07:00": compile(`%sT%s(Z|[+\-]\d{2}:\d{2})`, dateonly, datetime),
	"Mon, 02 Jan 2006 15:04:05 -0700":     compile(`(?i)[a-z]{3}, \d{2} (?i)[a-z]{3} \d{4} %s %s`, datetime, z0700),
	"Monday, 02-Jan-06 15:04:05 MST":      compile(`(?i)(Mon|Tues|Wednes|Thurs|Fri|Satur|Sun)day, \d{2}-(?i)[a-z]{3}-\d{2} %s %s`, datetime, mst),
	"Mon Jan 02 15:04:05 -0700 2006":      compile(`(?i)([a-z]{3} ){2}\d{2} %s %s \d{4}`, z0700, datetime),
	"Mon, 02 Jan 2006 15:04:05 MST":       compile(`(?i)[a-z]{3}, \d{2} (?i)[a-z]{3} \d{4} %s %s`, datetime, mst),
	"Mon Jan _2 15:04:05 MST 2006":        compile(`(?i)([a-z]{3} ){2}\d{1,2} %s %s \d{4}`, datetime, mst),
	"01/02 03:04:05PM '06 -0700":          compile(`\d{2}/\d{2} %s[AP]M '\d{2} %s`, datetime, z0700),
	"Mon Jan _2 15:04:05 2006":            compile(`(?i)([a-z]{3} ){2}\d{1,2} %s \d{4}`, datetime),
	"02 Jan 06 15:04 -0700":               compile(`\d{2} (?i)[a-z]{3} \d{2} \d{2}:\d{2} %s`, z0700),
	"02 Jan 06 15:04 MST":                 compile(`\d{2} (?i)[a-z]{3} \d{2} \d{2}:\d{2} %s`, mst),
	"2006-01-02 15:04:05":                 compile(dateonly + " " + datetime),
	"2006-01-02 15:04":                    compile(dateonly + ` \d{2}:\d{2}`),
	"Jan _2 15:04:05":                     compile(`(?i)[a-z]{3} \d{1,2} ` + datetime),
	"2006-01-02 15":                       compile(dateonly + ` \d{2}`),
	"2006-01-02":                          compile(dateonly),
	"15:04:05":                            compile(datetime),
	"2006-01":                             compile(`\d{4}-\d{2}`),
	"3:04PM":                              compile(`\d{1,2}:\d{2}[AP]M`),
	"15:04":                               compile(`\d{2}:\d{2}`),
	"2006":                                compile(`\d{4}`),
}

// ParseE 解析 value 并返回它所表示的时间
func ParseE(value string, loc ...*time.Location) (Time, error) {
	var layout string
	value = strings.Trim(strings.TrimSpace(value), `"`)

	if value == "" || value == `""` {
		return Time{}, nil
	}

	for k, v := range patterns {
		if v.MatchString(value) {
			layout = k
			break
		}
	}

	if loc == nil {
		loc = append(loc, time.Local)
	}

	pt, err := time.ParseInLocation(layout, value, loc[0])
	return Time{pt}, err
}

// Parse 返回忽略错误的 ParseE()
func Parse(value string, loc ...*time.Location) Time {
	t, _ := ParseE(value, loc...)
	return t
}

func compile(s string, a ...any) *regexp.Regexp {
	return regexp.MustCompile("^" + fmt.Sprintf(s, a...) + "$")
}
