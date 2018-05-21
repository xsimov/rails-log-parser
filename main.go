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
	isError                 bool
	timestamp               string
	PID, status             int
	exception, errorLines   string
	page, level, incomingIP string
	ARtime, viewTime        int
}

func main() {
	f, _ := ioutil.ReadFile("assets/small_production.log")
	scanner := bufio.NewScanner(bytes.NewReader(f))
	scanner.Split(logEntrySplit)
	for scanner.Scan() {
		e := parseLogEntry(scanner.Text())
		err := publishToES(e)
		if err != nil {
			log.Fatalf("could not publish to Elastic Search: %v", err)
		}
	}
}

func parseLogEntry(stringEntry string) (e logEntry) {
	IPRegexp := regexp.MustCompile(`\[(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d*)\s#(\d+)\]`)
	timeData := IPRegexp.FindAllStringSubmatch(stringEntry, -1)[0]
	e.timestamp = timeData[1]
	e.PID, _ = strconv.Atoi(timeData[2])
	fmt.Println(e)
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

func logEntrySplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	var errorLog []byte
	var errorAdvance int
	for {
		advance, token, err = bufio.ScanLines(data, atEOF)
		if err != nil {
			return 0, nil, err
		}
		if matched, regexpErr := regexp.Match("[I|D],", token); regexpErr != nil || matched || token == nil {
			if errorLog != nil {
				return errorAdvance, errorLog, nil
			}
			return
		}
		data = data[advance:]
		errorAdvance += advance
		errorLog = append(errorLog, token...)
		errorLog = append(errorLog, '\n')
	}
}

func (e logEntry) toJSON() ([]byte, error) {
	return json.Marshal(e)
}
