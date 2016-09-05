package impressive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/colm2/ical" // fork of soh335/ical that adds LOCATION
)

var (
	fs1 = "https://adfs.ucc.ie/adfs/ls/?SAMLRequest=jZHLboMwEEV%2fBXkfMI8iGAFS2i4aKVFRIF105zhOYwls6hmqfn4haaR0E3V57bnymeMCRd8NsBzpZLbqc1RI3nffGfTni5KNzoAVqBGM6BUCSWiWmzVEPgeBqBxpa5i3ei6ZPsgsOSaSh3EcHxMuZZ4%2b8CznWZTmccj3IfPelMOpULKpP7UQR7UySMLQdMTDdMHzBY%2faMIQoAp69M692lqy03aM2B20%2b7hPtL0MIL21bL%2brXpmXe8kr5ZA2OvXKNcl9aqt12XbIT0YAQBG4vJgM%2bSq0MaeFL2wezkzlJMbeDZvJRWyR28QNncfdxhl92VhXzNJz3dTf9%2f%2fqt7nFuFImDIHEmvIYiuHmxuqS%2f31z9AA%3d%3d&RelayState=I1D7dQS_Lu9_YUhHis9f-p2xJr2LKLco4FTR5_Gk9WYkfqBRhlTNNHChuP8AvvjfSY7HD7D-f3L6koBNJDNet_szRMorComquhoLDudi6fFeN3yFh_4hVLSj7zoqccbOSTjn5_-Wiz4CFSs6FBgHUAQQTC4nY_ZFF-RWtj1mJBj03JvPIFIMQjC1CeD1h2M_YUapZlTj-QycdBs6czd5Di2f7SuI8FJqlFfajDigBLldWYmj9VlnSRptQ0CqJ_lUssA12nMDqKbRr6q71ctwlCkcsFvzEmzWiN0LIQ-7joQETnn5-qwMp8z2y-G4XjVn2UQuMXjesZzjiFLciNkRnd3KfBIUKD0dgVhrlsQOm6PeKMj3KLb5BUnG5vqk4l4h8pRqWqrLyT0oC8I0Gqt__7tQTleNib1UE-EOWChSn00LcWQLFzvTLcVK0PJHa2pK"
	fs2 = "https://rbauth.scientia.com/Authentication/SamlPost"
	cal = "https://scientia-api-1-4-0.azurewebsites.net//api/Calendar"
)

const (
	timeFormat = "2006-01-02T15:04:05-07:00" //ISO8601
)

type CalendarResponse []RespObject

type RespObject struct {
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

func GetICal(email, password string) (string, error) {
	events, err := GetEvents(email, password)

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
		summary := fmt.Sprintf("%v %v", i.Name, i.EventType)
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

func GetEvents(email, password string) ([]ResourceEvent, error) {
	j, err := getCalendarJSONString(email, password)
	if err != nil {
		return nil, err
	}

	var resp CalendarResponse
	json.Unmarshal(j, &resp)

	return resp[0].ResourceEvents, nil
}

func getCalendarJSONString(email, password string) ([]byte, error) {
	token, err := fullGetToken(email, password)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", cal, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)

	return bytes, nil
}

func fullGetToken(email, password string) (string, error) {
	jar, _ := cookiejar.New(nil)

	client := &http.Client{Jar: jar}
	v1 := url.Values{}
	v1.Set("UserName", email)
	v1.Add("Password", password)
	v1.Add("AuthMethod", "FormsAuthentication")

	req1, err := http.NewRequest("POST", fs1, strings.NewReader(v1.Encode()))
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req1)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	b1 := string(bodyBytes)

	SAMLResponse, RelayState := divideFormInputs(b1)

	v2 := url.Values{}
	v2.Set("SAMLResponse", SAMLResponse)
	v2.Add("RelayState", RelayState)

	client2 := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req2, err := http.NewRequest("POST", fs2, strings.NewReader(v2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return "", err
	}
	resp2, err := client2.Do(req2)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	req3, err := http.NewRequest("GET", resp2.Header["Location"][0], nil)
	if err != nil {
		return "", err
	}

	resp3, err := client2.Do(req3)
	if err != nil {
		return "", err
	}
	defer resp3.Body.Close()

	return getTokenFromUrl(resp3.Header["Location"][0]), nil
}

func divideFormInputs(s1 string) (string, string) {
	sa := strings.SplitAfterN(s1, "value=\"", 2)
	sb := strings.SplitN(sa[1], "\"", 2)

	sc := strings.SplitAfterN(sb[1], "value=\"", 2)
	sd := strings.SplitN(sc[1], "\"", 2)

	return sb[0], sd[0]
}

func getTokenFromUrl(s2 string) string {
	sa := strings.SplitAfter(s2, "access_token=")
	sb := strings.Split(sa[1], "&")

	return "Bearer " + sb[0]
}
