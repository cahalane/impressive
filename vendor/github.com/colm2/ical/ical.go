package ical

// https://www.ietf.org/rfc/rfc2445.txt

import (
	"bufio"
	"io"
	"time"
)

const (
	stampLayout    = "20060102T150405Z"
	dateLayout     = "20060102"
	dateTimeLayout = "20060102T150405"
)

/*
  http://tools.ietf.org/html/draft-daboo-icalendar-extensions-09
  http://stackoverflow.com/a/17187346/195141

  BEGIN:VCALENDAR
  VERSION:2.0
  PRODID:-//My Company//NONSGML Event Calendar//EN
  URL:http://my.calendar/url
  NAME:My Calendar Name
  X-WR-CALNAME:My Calendar Name
  DESCRIPTION:A description of my calendar
  X-WR-CALDESC:A description of my calendar
  TIMEZONE-ID:Europe/London
  X-WR-TIMEZONE:Europe/London
  REFRESH-INTERVAL;VALUE=DURATION:PT12H
  X-PUBLISHED-TTL:PT12H
  COLOR:34:50:105
  CALSCALE:GREGORIAN
  METHOD:PUBLISH
*/

type VCalendar struct {
	VERSION string // 2.0
	PRODID  string // -//My Company//NONSGML Event Calendar//EN
	URL     string // http://my.calendar/url

	NAME         string // My Calendar Name
	X_WR_CALNAME string // My Calendar Name
	DESCRIPTION  string // A description of my calendar
	X_WR_CALDESC string // A description of my calendar

	TIMEZONE_ID   string // Europe/London
	X_WR_TIMEZONE string // Europe/London

	REFRESH_INTERVAL string // PT12H
	X_PUBLISHED_TTL  string // PT12H

	COLOR    string // 34:50:105
	CALSCALE string // GREGORIAN
	METHOD   string // PUBLISH

	VComponent []VComponent
}

func NewBasicVCalendar() *VCalendar {
	return &VCalendar{
		VERSION:  "2.0",
		CALSCALE: "GREGORIAN",
	}
}

func (c *VCalendar) Encode(w io.Writer) error {
	var b = bufio.NewWriter(w)

	if _, err := b.WriteString("BEGIN:VCALENDAR\r\n"); err != nil {
		return err
	}

	// use a slice map to preserve order during for range
	attrs := []map[string]string{
		{"VERSION:": c.VERSION},
		{"PRODID:": c.PRODID},
		{"URL:": c.URL},
		{"NAME:": c.NAME},
		{"X-WR-CALNAME:": c.X_WR_CALNAME},
		{"DESCRIPTION:": c.DESCRIPTION},
		{"X-WR-CALDESC:": c.X_WR_CALDESC},
		{"TIMEZONE-ID:": c.TIMEZONE_ID},
		{"X-WR-TIMEZONE:": c.X_WR_TIMEZONE},
		{"REFRESH-INTERVAL;VALUE=DURATION:": c.REFRESH_INTERVAL},
		{"X-PUBLISHED-TTL:": c.X_PUBLISHED_TTL},
		{"COLOR:": c.COLOR},
		{"CALSCALE:": c.CALSCALE},
		{"METHOD:": c.METHOD},
	}

	for _, item := range attrs {
		for k, v := range item {
			if len(v) == 0 {
				continue
			}
			if _, err := b.WriteString(k + v + "\r\n"); err != nil {
				return err
			}
		}
	}

	for _, component := range c.VComponent {
		if err := component.EncodeIcal(b); err != nil {
			return err
		}
	}

	if _, err := b.WriteString("END:VCALENDAR\r\n"); err != nil {
		return err
	}

	return b.Flush()
}

type VComponent interface {
	EncodeIcal(w io.Writer) error
}

type VEvent struct {
	UID         string
	DTSTAMP     time.Time
	DTSTART     time.Time
	DTEND       time.Time
	SUMMARY     string
	DESCRIPTION string
	TZID        string
	LOCATION    string

	AllDay bool
}

func (e *VEvent) EncodeIcal(w io.Writer) error {

	var timeStampLayout, timeStampType, tzidTxt string

	if e.AllDay {
		timeStampLayout = dateLayout
		timeStampType = "DATE"
	} else {
		timeStampLayout = dateTimeLayout
		timeStampType = "DATE-TIME"
		if len(e.TZID) == 0 || e.TZID == "UTC" {
			timeStampLayout = timeStampLayout + "Z"
		}
	}

	if len(e.TZID) != 0 && e.TZID != "UTC" {
		tzidTxt = "TZID=" + e.TZID + ";"
	}

	b := bufio.NewWriter(w)
	if _, err := b.WriteString("BEGIN:VEVENT\r\n"); err != nil {
		return err
	}
	if _, err := b.WriteString("DTSTAMP:" + e.DTSTAMP.UTC().Format(stampLayout) + "\r\n"); err != nil {
		return err
	}
	if _, err := b.WriteString("UID:" + e.UID + "\r\n"); err != nil {
		return err
	}

	if len(e.TZID) != 0 && e.TZID != "UTC" {
		if _, err := b.WriteString("TZID:" + e.TZID + "\r\n"); err != nil {
			return err
		}
	}

	if _, err := b.WriteString("SUMMARY:" + e.SUMMARY + "\r\n"); err != nil {
		return err
	}
	if e.DESCRIPTION != "" {
		if _, err := b.WriteString("DESCRIPTION:" + e.DESCRIPTION + "\r\n"); err != nil {
			return err
		}
	}
	if e.LOCATION != "" {
		if _, err := b.WriteString("LOCATION:" + e.LOCATION + "\r\n"); err != nil {
			return err
		}
	}
	if _, err := b.WriteString("DTSTART;" + tzidTxt + "VALUE=" + timeStampType + ":" + e.DTSTART.Format(timeStampLayout) + "\r\n"); err != nil {
		return err
	}

	if _, err := b.WriteString("DTEND;" + tzidTxt + "VALUE=" + timeStampType + ":" + e.DTEND.Format(timeStampLayout) + "\r\n"); err != nil {
		return err
	}

	if _, err := b.WriteString("END:VEVENT\r\n"); err != nil {
		return err
	}

	return b.Flush()
}
