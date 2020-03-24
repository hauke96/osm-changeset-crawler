package analysis

import (
	"strings"
	"sync"
	"time"

	"github.com/hauke96/osm-changeset-crawler/common"
	"github.com/hauke96/sigolo"
)

// analyseEditorCount takes bunches of "aggregationSize" many changesets and
// counts their edits. The result is written to the given file in a CSV format.
func AnalyseUserWithoutSource(outputPath string, changsetChannel <-chan []common.Changeset, finishWaitGroup *sync.WaitGroup) {
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

			// ID 0 indicates an empty cache place
			if changeset.Id == 0 || changeset.Source != "" {
				continue
			}

			editor := common.EDITOR_NOT_SET
			user := changeset.User

			if _, ok := aggregationMap[user]; !ok {
				aggregationMap[user] = make(map[string]int)
			}

			createdBy := strings.ToLower(changeset.CreatedBy)
			for _, e := range common.KNOWN_EDITORS {
				if strings.Contains(createdBy, e) {
					sigolo.Debug("Editor found: %s", e)
					editor = e
					break
				}
			}

			aggregationMap[user][editor]++
		}

		sigolo.Info("Counted values for %d users, which took %dms", common.CACHE_SIZE, time.Since(clock).Milliseconds())
	}

	writeToFile(columnCount, common.KNOWN_EDITORS, true, aggregationMap, writer)
}
