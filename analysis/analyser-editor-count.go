package analysis

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hauke96/osm-changeset-analyser/common"
	"github.com/hauke96/sigolo"
)

// analyseEditorCount takes bunches of "aggregationSize" many changesets and
// counts their edits. The result is written to the given file in a CSV format.
func AnalyseEditorCount(outputPath string, changsetChannel <-chan []common.Changeset) {
	clock := time.Now()
	// columnCount is the amount of column in the CSV file. The value
	// "len(knownEditors)+1" is the mount of all editors plus column for
	// changeset count
	columnCount := len(common.KNOWN_EDITORS) + 1
	aggregationMap := make(map[string]map[string]int)
	// writtenAggregations is the number of lines in the CSV file. This is used
	// to increase the value of the first column showing the amount of
	// changesets for that row
	receivedChangesetSets := 0

	var currentCreatedAt string

	// Open CSV and create writer
	file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0644)
	sigolo.FatalCheck(err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write first head line with the column names
	headLine := make([]string, columnCount)
	headLine[0] = "changeset count"
	for i := 0; i < len(common.KNOWN_EDITORS); i++ {
		headLine[i+1] = common.KNOWN_EDITORS[i]
	}

	err = writer.Write(headLine)
	sigolo.FatalCheck(err)
	writer.Flush()

	// Go through the changesets and calculate the amount of editor per
	// "aggregationSize" many changesets
	for changesets := range changsetChannel {
		sigolo.Info("Received changesets set %d -> count editors", receivedChangesetSets)
		receivedChangesetSets++

		clock = time.Now()

		for _, changeset := range changesets {
			sigolo.Debug("Look at changeset %#v", changeset)

			// ID 0 inidcates an empty cache place
			if changeset.Id == 0 || changeset.CreatedAt == "" {
				continue
			}

			editor := common.EDITOR_NOT_SET
			createdAt := changeset.CreatedAt[0:7] // e.g. "2020-04"

			if _, ok := aggregationMap[createdAt]; !ok {
				sigolo.Info("Create new map for '%s'", createdAt)
				aggregationMap[createdAt] = make(map[string]int)
			}

			createdBy := strings.ToLower(changeset.CreatedBy)
			for _, e := range common.KNOWN_EDITORS {
				if strings.Contains(createdBy, e) {
					sigolo.Debug("Editor found: %s", e)
					editor = e
					break
				}
			}

			aggregationMap[createdAt][editor]++
		}

		sigolo.Info("Counted %d editors which took %dms", common.CACHE_SIZE, time.Since(clock).Milliseconds())
	}

	writeToFile(columnCount, currentCreatedAt, aggregationMap, writer)
}

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
