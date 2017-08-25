package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/colm2/ical" // fork of soh335/ical that adds LOCATION
)

var (
	mods = make(map[string]string)
)

const (
	timeFormat = "2006-01-02T15:04:05-07:00" //ISO8601
)

type calendarResponse []respObject

type respObject struct {
	Identity       string          `json:"Identity"`
	ResourceEvents []ResourceEvent `json:"ResourceEvents"`
}

type ResourceEvent struct {
	Identity      string `json:"Identity"`
	Name          string `json:"Name"`
	Location      string `json:"Location"`
	EventType     string `json:"EventType"`
	StartDateTime string `json:"StartDateTime"`
	EndDateTime   string `json:"EndDateTime"`
}

type retrieveCal struct{}

func main() {
	portF := flag.String("port", ":3000", "Port number of server, default \":3000\"")
	http.Handle("/", &retrieveCal{})

	log.Fatal(http.ListenAndServe(*portF, nil))
}

func (i *retrieveCal) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)

	cal, err := GetICal(body)

	if err != nil {
		io.WriteString(resp, "{\"ok\":false, \"error\":\""+err.Error()+"\"}")
	} else {
		resp.Header().Set("Content-Type", "text/calendar")
		io.WriteString(resp, cal)
	}
}

func GetICal(j []byte) (str string, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	events, err := GetEvents(j)

	c := ical.NewBasicVCalendar()
	c.PRODID = "UCC Timetable"
	c.TIMEZONE_ID = "Europe/Dublin"
	c.NAME = "UCC Timetable"
	c.DESCRIPTION = "Timetable from mytimetable.ucc.ie extracted by Impressive"
	c.COLOR = "116:180:80" // UCC's preferred green, actually

	if err != nil {
		return "", err
	}

	for _, i := range events {
		start, err := time.Parse(timeFormat, i.StartDateTime)
		if err != nil {
			return "", err
		}
		end, err := time.Parse(timeFormat, i.EndDateTime)
		if err != nil {
			return "", err
		}

		var summary string
		modName, found := mods[strings.Split(i.Name, "/")[0]]
		if found {
			summary = fmt.Sprintf("%v %v (%v)", i.Name, i.EventType, modName)
		} else {
			summary = fmt.Sprintf("%v %v", i.Name, i.EventType)
		}

		e := &ical.VEvent{
			UID:      i.Identity,
			DTSTART:  start,
			DTEND:    end,
			SUMMARY:  summary,
			LOCATION: i.Location,
		}

		c.VComponent = append(c.VComponent, e)
	}

	var b bytes.Buffer
	if err := c.Encode(&b); err != nil {
		return "", err
	}

	return b.String(), err
}

func GetEvents(j []byte) ([]ResourceEvent, error) {
	var resp calendarResponse
	json.Unmarshal(j, &resp)

	return resp[0].ResourceEvents, nil
}
