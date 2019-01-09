// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

var (
	est *time.Location
)

func init() {
	// US Eastern Time Zone
	est, _ = time.LoadLocation("America/New_York")
}

func TestTime(t *testing.T) {
	t1 := Now()

	// Verify time.Time methods work
	if diff := t1.Sub(t1.Time); diff != 0 {
		t.Errorf("got %v", diff)
	}
	if tt := time.Now().Add(1 * time.Second); t1.Sub(tt) == 0 {
		t.Error("expected difference in timing")
	}
}

func TestTime__NewTime(t *testing.T) {
	f := func(_ Time) {}
	f(NewTime(time.Now())) // make sure we can lift time.Time values

	start := time.Now().Add(-1 * time.Second)

	// Example from NewTime godoc
	now := Now()
	fmt.Println(start.Sub(now.Time))
}

func TestTime__Negative(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	// ts is 000000 (yymmdd) and 0000 (hhmm) from an ACH file,
	// then we convert to NYC timezone
	ts := time.Date(0, time.January, 0, 0, 0, 0, 0, time.UTC).In(loc)

	if !ts.Before(time.Time{}) {
		// ts should be negative now (i.e. -0001-12-30 19:03:58 -0456 LMT), which is bogus
		t.Errorf("%s isn't negative..", ts.String())
	}

	tt := NewTime(ts) // wrap, which should fix our problem
	if !tt.IsZero() {
		t.Errorf("expected tt to be zero time: %v", tt.String())
	}
	if tt.Before(time.Time{}) {
		t.Errorf("tt shouldn't be before zero time: %v", tt.String())
	}
}

func TestTime__JSON(t *testing.T) {
	// marshal and then unmarshal
	t1 := Now()

	bs, err := t1.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var t2 Time
	if err := json.Unmarshal(bs, &t2); err != nil {
		t.Fatal(err)
	}
	if !t1.Equal(t2) {
		t.Errorf("unequal: t1=%q t2=%q", t1, t2)
	}

	in := []byte(`"2018-11-27T00:54:53Z"`)
	var t3 Time
	if err := json.Unmarshal(in, &t3); err != nil {
		t.Fatal(err)
	}
	if t3.IsZero() {
		t.Error("t3 shouldn't be zero time")
	}

	// empty should unmarshal to nothing
	in = []byte(`""`)
	var t4 Time
	if err := json.Unmarshal(in, &t4); err != nil {
		t.Errorf("empty value for base.Time is fine, but got: %v", err)
	}
}

func TestTime__jsonRFC3339(t *testing.T) {
	// Read RFC 3339 time
	in := []byte(fmt.Sprintf(`"%s"`, time.Now().Format(time.RFC3339)))
	var t1 Time
	if err := json.Unmarshal(in, &t1); err != nil {
		t.Fatal(err)
	}
	if t1.IsZero() {
		t.Error("t4 shouldn't be zero time")
	}
}

func TestTime__javascript(t *testing.T) {
	// Generated with (new Date).toISOString() in Chrome and Firefox
	in := []byte(`{"time": "2018-12-14T20:36:58.789Z"}`)

	type wrapper struct {
		When Time `json:"time"`
	}
	var wrap wrapper
	if err := json.Unmarshal(in, &wrap); err != nil {
		t.Fatal(err)
	}
	if v := wrap.When.String(); v != "2018-12-14 15:36:58 -0500 EST" {
		t.Errorf("got %q", v)
	}
}

var quote = []byte(`"`)

// TestTime__ruby will attempt to parse an ISO 8601 time generated by this library
func TestTime__ruby(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping ruby ISO 8601 test on windows")
	}

	bin, err := exec.LookPath("ruby")
	if err != nil || bin == "" {
		if inCI := os.Getenv("TRAVIS_OS_NAME") != ""; inCI {
			t.Fatal("ruby not found")
		} else {
			t.Skip("ruby not found")
		}
	}

	tt, err := time.Parse(iso8601Format, "2018-11-18T09:04:23-08:00")
	if err != nil {
		t.Fatal(err)
	}
	t1 := Time{
		Time: tt,
	}

	bs, err := t1.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	bs = bytes.TrimPrefix(bytes.TrimSuffix(bs, quote), quote)

	// Check with ruby
	cmd := exec.Command(bin, "time.rb", string(bs))
	bs, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("err=%v\nOutput: %v", err, string(bs))
	}

	// Validate ruby output
	if !bytes.Contains(bs, []byte(`Date: 2018-11-18`)) {
		t.Errorf("no Date: %v", string(bs))
	}
	if !bytes.Contains(bs, []byte(`Time: 09:04:23`)) {
		t.Errorf("no Time: %v", string(bs))
	}
}

func TestTime__IsBankingDay(t *testing.T) {
	tests := []struct {
		Date     time.Time
		Expected bool
	}{
		// new years day
		{time.Date(2018, time.January, 1, 1, 0, 0, 0, est), false},
		// Wednesday Canary test
		{time.Date(2018, time.January, 3, 1, 0, 0, 0, est), true},
		// saturday
		{time.Date(2018, time.January, 6, 1, 0, 0, 0, est), false},
		// sunday
		{time.Date(2018, time.January, 7, 1, 0, 0, 0, est), false},
		// Martin Luther King, JR. Day
		{time.Date(2018, time.January, 15, 1, 0, 0, 0, est), false},
		// Presidents' Day
		{time.Date(2018, time.February, 19, 1, 0, 0, 0, est), false},
		// Memorial Day
		{time.Date(2018, time.May, 28, 1, 0, 0, 0, est), false},
		// Independence Day
		{time.Date(2018, time.July, 4, 1, 0, 0, 0, est), false},
		// Labor Day
		{time.Date(2018, time.September, 3, 1, 0, 0, 0, est), false},
		// Columbus Day
		{time.Date(2018, time.October, 8, 1, 0, 0, 0, est), false},
		// Vesterans' Day Observed on the monday
		{time.Date(2018, time.November, 12, 1, 0, 0, 0, est), false},
		// Thanksgiving Day
		{time.Date(2018, time.November, 22, 1, 0, 0, 0, est), false},
		// Christmas Day
		{time.Date(2018, time.December, 25, 1, 0, 0, 0, est), false},
	}
	for _, test := range tests {
		actual := NewTime(test.Date).IsBankingDay()
		if actual != test.Expected {
			t.Errorf("Date %s: expected %t, got %t", test.Date, test.Expected, actual)
		}

		actual = NewTime(test.Date).IsBankingDay()
		if actual != test.Expected {
			t.Errorf("Date %s: expected %t, got %t", test.Date, test.Expected, actual)
		}
	}
}

func TestTime__IsWeekend(t *testing.T) {
	tests := []struct {
		Date     time.Time
		Expected bool
	}{
		// saturday
		{time.Date(2018, time.January, 6, 1, 0, 0, 0, est), true},
		// sunday
		{time.Date(2018, time.January, 7, 1, 0, 0, 0, est), true},
		// monday
		{time.Date(2018, time.January, 9, 1, 0, 0, 0, est), false},
	}
	for _, test := range tests {
		actual := NewTime(test.Date).IsWeekend()
		if actual != test.Expected {
			t.Errorf("Date %s: expected %t, got %t", test.Date, test.Expected, actual)
		}

		actual = NewTime(test.Date).IsWeekend()
		if actual != test.Expected {
			t.Errorf("Date %s: expected %t, got %t", test.Date, test.Expected, actual)
		}
	}
}

func TestTime_AddBankingDay(t *testing.T) {
	tests := []struct {
		Date   time.Time
		Future time.Time
		Days   int
	}{
		// Thursday add two days over a monday holiday abd needs to be following tuesday
		{time.Date(2018, time.January, 11, 1, 0, 0, 0, est), time.Date(2018, time.January, 16, 1, 0, 0, 0, est), 2},
	}
	for _, test := range tests {
		actual := NewTime(test.Date).AddBankingDay(test.Days)
		if !actual.Equal(NewTime(test.Future)) {
			t.Errorf("Adding %d days: expected %s, got %s", test.Days, test.Future.Weekday().String(), actual)
		}

		actual = NewTime(test.Date).AddBankingDay(test.Days)
		if !actual.Equal(NewTime(test.Future)) {
			t.Errorf("Adding %d days: expected %s, got %s", test.Days, test.Future.Weekday().String(), actual)
		}
	}
}

func TestTime__Conversions(t *testing.T) {
	// create dates that are on a different day earlier than a holiday in different time zone
	pacific, _ := time.LoadLocation("America/Los_Angeles")
	when := NewTime(time.Date(2018, time.December, 24, 23, 0, 0, 0, pacific))
	if when.Day() != 25 {
		t.Errorf("%v but expected to fall on Christmas", t)
	}

	// create dates that are on a different day later than a holiday in different time zone
	madrid, _ := time.LoadLocation("Europe/Madrid")
	when = NewTime(time.Date(2018, time.December, 26, 0, 30, 0, 0, madrid))
	if when.Day() != 25 {
		t.Errorf("%v but expected to fall on Christmas", t)
	}
}
