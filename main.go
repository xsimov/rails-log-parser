package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

type logEntry struct {
	isError                 bool
	timestamp               string
	PID, status             int
	exception, lines        string
	page, level, incomingIP string
	ARtime, viewTime        int
}

func main() {
	f, _ := ioutil.ReadFile("assets/small_production.log")
	scanner := bufio.NewScanner(bytes.NewReader(f))
	scanner.Split(logEntrySplit)
	for scanner.Scan() {
		t := scanner.Text()
		e := parseLogEntry(t)
		// err := publishToES(e)
		// if err != nil {
		// 	log.Fatalf("could not publish to Elastic Search: %v", err)
		// }
		fmt.Println(t, e, "---LOG LINE---")
	}
}

func parseLogEntry(stringEntry string) (e logEntry) {
	IPRegexp := regexp.MustCompile(`\[(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d*)\s#(\d+)\]`)
	timeData := IPRegexp.FindAllStringSubmatch(stringEntry, -1)[0]
	e.timestamp = timeData[1]
	e.PID, _ = strconv.Atoi(timeData[2])
	return
}

func publishToES(e logEntry) error {
	jsonEntry, err := e.toJSON()
	if err != nil {
		return fmt.Errorf("could not marshal JSON: %v", err)
	}
	_, err = http.Post("http://localhost:9200/log_entries/_doc", "application/json", bytes.NewReader(jsonEntry))
	if err != nil {
		return fmt.Errorf("elasticsearch server is unreachable: %v", err)
	}
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
