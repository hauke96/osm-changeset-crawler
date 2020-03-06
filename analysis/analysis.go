package analysis

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/hauke96/osm-changeset-analyser/common"
	"github.com/hauke96/sigolo"
)

func writeToFile(columnCount int, keyColumnValue string, aggregationMap map[string]map[string]int, writer *csv.Writer) {
	sigolo.Debug("Write %#v", aggregationMap)
	line := make([]string, columnCount)

	month := 1
	year := 2000
	finalDateString := time.Now().Format("2006-01")

	for {
		dateString := fmt.Sprintf("%d-%02d", year, month)

		editorToCount := aggregationMap[dateString]

		line[0] = dateString
		i := 1
		for _, e := range common.KNOWN_EDITORS {
			line[i] = strconv.Itoa(editorToCount[e])
			i++
		}

		err := writer.Write(line)
		sigolo.FatalCheck(err)

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
