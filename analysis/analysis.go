package analysis

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hauke96/osm-changeset-crawler/common"
	"github.com/hauke96/sigolo"
)

// writeToFileWithDates writes the given data (aggregation map) sorted by month
// using the writer. The date format is "2006-01" and starts with "2000-01" and
// goes on until the current month.
// The "dataColumnNames" contains the column names for the head line of the data,
// which is basicall the had line without the first column (which contains the
// dates in this case).
// You can add an additional column with the sum of each row by setting
// "addTotalCountColumn" to true.
func writeToFileWithDates(columnCount int, dataColumnNames []string, addTotalCountColumn bool, aggregationMap map[string]map[string]int, writer *csv.Writer) {
	sigolo.Debug("Write %#v", aggregationMap)

	// First changeset was 9th April 2005
	month := 4
	year := 2005
	finalDateString := time.Now().Format("2006-01")

	for {
		dateString := fmt.Sprintf("%d-%02d", year, month)
		editorToCount := aggregationMap[dateString]

		if addTotalCountColumn {
			writeLineWithTotalCount(columnCount, dateString, dataColumnNames, editorToCount, writer)
		} else {
			writeLine(columnCount, dateString, dataColumnNames, editorToCount, writer)
		}

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

func writeToFile(columnCount int, dataColumnNames []string, addTotalCountColumn bool, aggregationMap map[string]map[string]int, writer *csv.Writer) {
	sigolo.Debug("Write %#v", aggregationMap)

	for dateString, editorToCount := range aggregationMap {
		if addTotalCountColumn {
			writeLineWithTotalCount(columnCount, dateString, dataColumnNames, editorToCount, writer)
		} else {
			writeLine(columnCount, dateString, dataColumnNames, editorToCount, writer)
		}
	}

	writer.Flush()
}

func writeLine(columnCount int, firstColumnName string, dataColumnNames []string, data map[string]int, writer *csv.Writer) {
	line := make([]string, columnCount)
	line[0] = firstColumnName
	i := 1

	for _, e := range dataColumnNames {
		line[i] = strconv.Itoa(data[e])
		i++
	}

	err := writer.Write(line)
	sigolo.FatalCheck(err)
}

func writeLineWithTotalCount(columnCount int, firstColumnName string, dataColumnNames []string, data map[string]int, writer *csv.Writer) {
	line := make([]string, columnCount+1) // +1 for the "all" column
	totalCount := 0
	line[0] = firstColumnName
	i := 1

	for _, e := range dataColumnNames {
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

func createHeadLine(firstColumn string, dataValues []string) []string {
	// Write first head line with the column names
	headLine := make([]string, len(dataValues)+1)
	headLine[0] = firstColumn
	for i := 0; i < len(dataValues); i++ {
		headLine[i+1] = dataValues[i]
	}

	return headLine
}
