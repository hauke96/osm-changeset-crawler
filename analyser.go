package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hauke96/sigolo"
)

// analyseEditorCount takes bunches of "aggregationSize" many changesets and
// counts their edits. The result is written to the given file in a CSV format.
func analyseEditorCount(outputPath string, changsetChannel <-chan []Changeset) {
	clock := time.Now()
	// columnCount is the amount of column in the CSV file. The value
	// "len(knownEditors)+1" is the mount of all editors plus column for
	// changeset count
	columnCount := len(knownEditors) + 1
	aggregationMap := make(map[string]int)
	processedChangesets := 0
	// writtenAggregations is the number of lines in the CSV file. This is used
	// to increase the value of the first column showing the amount of
	// changesets for that row
	writtenAggregations := 0
	receivedChangesetSets := 0

	// Open CSV and create writer
	file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0644)
	sigolo.FatalCheck(err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write first head line with the column names
	headLine := make([]string, columnCount)
	headLine[0] = "changeset count"
	for i := 0; i < len(knownEditors); i++ {
		headLine[i+1] = knownEditors[i]
	}

	err = writer.Write(headLine)
	sigolo.FatalCheck(err)

	// Go through the changesets and calculate the amount of editor per
	// "aggregationSize" many changesets
	for changesets := range changsetChannel {
		sigolo.Info("Received changesets set %d -> count editors", receivedChangesetSets)
		receivedChangesetSets++

		for _, changeset := range changesets {
			// ID 0 inidcates an empty cache place
			if changeset.Id == 0 {
				continue
			}

			editor := noEditor

			for _, tag := range changeset.Tags {
				if tag.K == editorKey {
					for _, e := range knownEditors {
						if strings.Contains(strings.ToLower(tag.V), e) {
							editor = e
							break
						}
					}
					break
				}
			}

			aggregationMap[editor]++
			processedChangesets++
		}

		sigolo.Info("Counted %d editors which took %dms", CACHE_SIZE, time.Since(clock).Milliseconds())
		clock = time.Now()

		writtenAggregations++
		processedChangesets = 0

		writeCountToFile(columnCount, CACHE_SIZE*writtenAggregations-1, aggregationMap, writer)

		aggregationMap = make(map[string]int)
	}

	if processedChangesets != 0 {
		writeCountToFile(columnCount, CACHE_SIZE*writtenAggregations+processedChangesets, aggregationMap, writer)
	}
}

func writeCountToFile(columnCount, processedChangesets int, aggregationMap map[string]int, writer *csv.Writer) {
	sigolo.Debug("Write %#v", aggregationMap)
	line := make([]string, columnCount)
	line[0] = strconv.Itoa(processedChangesets + 1)
	for i := 1; i < columnCount; i++ {
		line[i] = strconv.Itoa(aggregationMap[knownEditors[i-1]])
	}

	sigolo.Info("Write data for key '%s' to file", line[0])

	err := writer.Write(line)
	sigolo.FatalCheck(err)
}
