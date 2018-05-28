package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type logEntry struct {
	IsError    bool   `json:"isError"`
	Timestamp  string `json:"timestamp"`
	Method     string `json:"method"`
	Lines      string `json:"lines"`
	Path       string `json:"path"`
	IncomingIP string `json:"IP"`
	PID        int    `json:"PID"`
	Status     int    `json:"status"`
	ARtime     int    `json:"activeRecordTime"`
	ViewTime   int    `json:"viewTime"`
}

var (
	timestampRegexp = regexp.MustCompile(`\[(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d*)\s#(\d+)\]`)
	ipRegexp        = regexp.MustCompile(`Started (GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS) (.*) for (\d+\.\d+\.\d+\.\d+)`)
	levelRegexp     = regexp.MustCompile(`(?m)^F,\s`)
)

func main() {
	f, _ := ioutil.ReadFile("assets/small_production.log")
	scanner := bufio.NewScanner(bytes.NewReader(f))
	scanner.Split(logEntrySplit)
	for scanner.Scan() {
		t := scanner.Text()
		if t == "\n" {
			return
		}
		e := parseLogEntry(t)
		r, err := e.toJSON()
		fmt.Printf("%s, %v", r, err)
		err = publishToES(e)
		if err != nil {
			log.Fatalf("could not publish to Elastic Search: %v", err)
		}
	}
}

func parseLogEntry(stringEntry string) (e logEntry) {
	e.Lines = stringEntry
	timestamp, PID := getTimestampAndPID(stringEntry)
	e.Timestamp, e.PID = fmt.Sprintf("%sZ", timestamp), PID
	e.Method, e.Path, e.IncomingIP = getIP(stringEntry)
	e.IsError = levelRegexp.MatchString(stringEntry)
	if e.IsError {
		e.Status = 500
	} else {
		e.Status = 200
	}
	return
}

func getTimestampAndPID(s string) (string, int) {
	timeData := timestampRegexp.FindAllStringSubmatch(s, -1)[0]
	pid, _ := strconv.Atoi(timeData[2])
	return timeData[1], pid
}

func getIP(s string) (method, path, ip string) {
	data := ipRegexp.FindAllStringSubmatch(s, -1)
	if len(data) > 0 {
		d := data[0]
		return d[1], d[2], d[3]
	}
	return
}

func publishToES(e logEntry) error {
	jsonEntry, err := e.toJSON()
	if err != nil {
		return fmt.Errorf("could not marshal JSON: %v", err)
	}
	resp, err := http.Post("http://localhost:9200/log_entries/rails/", "application/json", bytes.NewReader(jsonEntry))
	if err != nil {
		return fmt.Errorf("elasticsearch server is unreachable: %v", err)
	}
	fmt.Println(resp)
	return nil
}

func logEntrySplit(data []byte, atEOF bool) (multilineAdvance int, logLines []byte, err error) {
	for {
		advance, token, err := bufio.ScanLines(data, atEOF)
		if err != nil {
			return 0, nil, err
		}
		if matched, regexpErr := regexp.Match(`\sStarted (GET|POST|PUT|PATCH|DELETE) "`, token); regexpErr != nil || matched || token == nil {
			if logLines != nil {
				return multilineAdvance, logLines, nil
			}
		}
		data = data[advance:]
		multilineAdvance += advance
		logLines = append(logLines, token...)
		logLines = append(logLines, '\n')
	}
}

func (e logEntry) toJSON() ([]byte, error) {
	return json.Marshal(e)
}
