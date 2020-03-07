package analysis

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hauke96/osm-changeset-analyser/common"
	"github.com/hauke96/sigolo"
)

func writeToFileWithDates(columnCount int, aggregationMap map[string]map[string]int, writer *csv.Writer) {
	sigolo.Debug("Write %#v", aggregationMap)

	month := 1
	year := 2000
	finalDateString := time.Now().Format("2006-01")

	for {
		dateString := fmt.Sprintf("%d-%02d", year, month)
		editorToCount := aggregationMap[dateString]

		writeLine(columnCount, dateString, editorToCount, writer)

		month++
		if month == 13 {
			month = 1
			year++
		}

		if dateString == finalDateString {
			break
		}
	}

	writer.Flush()
}

func writeToFile(columnCount int, aggregationMap map[string]map[string]int, writer *csv.Writer) {
	sigolo.Debug("Write %#v", aggregationMap)

	for k, v := range aggregationMap {
		writeLine(columnCount, k, v, writer)
	}

	writer.Flush()
}

func writeLine(columnCount int, key string, data map[string]int, writer *csv.Writer) {
	line := make([]string, columnCount+1) // +1 for the "all" column
	totalCount := 0
	line[0] = key
	i := 1

	for _, e := range common.KNOWN_EDITORS {
		totalCount += data[e]
		line[i] = strconv.Itoa(data[e])
		i++
	}

	line[i] = strconv.Itoa(totalCount)

	err := writer.Write(line)
	sigolo.FatalCheck(err)
}

func initAnalyser(outputPath string, headLine []string) (time.Time, map[string]map[string]int, *csv.Writer) {
	clock := time.Now()

	aggregationMap := make(map[string]map[string]int)
	// writtenAggregations is the number of lines in the CSV file. This is used
	// to increase the value of the first column showing the amount of
	// changesets for that row

	// Open CSV and create writer
	file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0644)
	sigolo.FatalCheck(err)

	writer := csv.NewWriter(file)
	err = writer.Write(headLine)
	sigolo.FatalCheck(err)

	writer.Flush()
	return clock, aggregationMap, writer
}

func createEditorHeadLine(columnCount int) []string {
	// Write first head line with the column names
	headLine := make([]string, columnCount)
	headLine[0] = "changeset count"
	for i := 0; i < len(common.KNOWN_EDITORS); i++ {
		headLine[i+1] = common.KNOWN_EDITORS[i]
	}

	return headLine
}
