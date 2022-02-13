package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/apognu/gocal"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseTimezone(s string) *time.Location {
	tz, err := time.LoadLocation(s)
	if err == nil {
		return tz
	}

	parts := strings.Split(s, " ")
	if len(parts) == 0 {
		return nil
	}

	newTz := ""
	for _, word := range parts {
		newTz += string(word[0])
	}

	tz, err = time.LoadLocation(newTz)
	if err == nil {
		return tz
	}

	return nil
}

func newTime(t *time.Time, tz *time.Location) *time.Time {
	new := time.Date(t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
	return &new
}

func run() error {
	if len(os.Args) != 2 {
		return errors.New("missing ics file")
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}

	c := gocal.NewParser(file)
	c.SkipBounds = true
	if err := c.Parse(); err != nil {
		return err
	}

	for _, e := range c.Events {
		rawTz, ok := e.RawStart.Params["TZID"]
		if ok {
			tz := parseTimezone(rawTz)
			if tz != nil {
				e.Start = newTime(e.Start, tz)
				e.End = newTime(e.End, tz)
			}
		}
		fmt.Printf("Event id:\t%s\n", e.Uid)
		fmt.Printf("Start:\t\t%s\n", e.Start.In(time.Local).Format(time.UnixDate))
		fmt.Printf("End:\t\t%s\n", e.End.In(time.Local).Format(time.UnixDate))
		if e.Duration == nil {
			duration := e.End.Sub(*e.Start)
			e.Duration = &duration
		}
		fmt.Printf("Duration:\t%s\n", e.Duration)
		fmt.Printf("Location:\t%s\n", e.Location)

		fmt.Printf("Attendees:\n")
		for i, attendee := range e.Attendees {
			if i == 5 {
				fmt.Printf("\t\tand %d more...\n", len(e.Attendees)-i)
				break
			}
			fmt.Printf("\t\t%s\n", attendee.Cn)
		}

		if e.Summary != "" {
			fmt.Printf("Summary:\t%s\n", e.Summary)
		}

		if e.Organizer != nil {
			fmt.Printf("Organizer:\t%s\n", e.Organizer.Cn)
		}

		if e.Description != "" {
			fmt.Printf("Description:\n%s\n", strings.ReplaceAll(e.Description, "\\n", "\n"))
		}
	}

	return nil
}
