package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"time"
)

type logEntry struct {
	level     string
	isError   bool
	timestamp time.Time
	PID       int
}

func main() {
	f, _ := ioutil.ReadFile("assets/small_production.log")
	scanner := bufio.NewScanner(bytes.NewReader(f))
	scanner.Split(logEntrySplit)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
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
