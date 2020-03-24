package analysis

import (
	"strings"
	"sync"
	"time"

	"github.com/hauke96/osm-changeset-analyser/common"
	"github.com/hauke96/sigolo"
)

// analyseEditorCount takes bunches of "aggregationSize" many changesets and
// counts their edits. The result is written to the given file in a CSV format.
func AnalyseEditorCount(outputPath string, changsetChannel <-chan []common.Changeset, finishWaitGroup *sync.WaitGroup) {
	defer finishWaitGroup.Done()

	// columnCount is the amount of column in the CSV file. The value
	// "len(knownEditors)+1" is the mount of all editors plus column for
	// changeset count
	columnCount := len(common.KNOWN_EDITORS) + 1

	headLine := createEditorHeadLine(columnCount)

	clock, aggregationMap, writer := initAnalyser(outputPath, headLine)

	// Go through the changesets and calculate the amount of editor per
	// "aggregationSize" many changesets
	for changesets := range changsetChannel {
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

	writeToFileWithDates(columnCount, common.KNOWN_EDITORS, aggregationMap, writer)
}
