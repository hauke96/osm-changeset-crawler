package analysis

import (
	"strings"
	"sync"
	"time"

	"github.com/hauke96/osm-changeset-analyser/common"
	"github.com/hauke96/sigolo"
)

// AnalyseCommentKeywordCount searches for the given keywords in the changesets
// after converting everythig to lowercase. In the end the keywords are the
// different columns of the CSV file.
func AnalyseCommentKeywordsCount(outputPath string, keywords []string, changsetChannel <-chan []common.Changeset, finishWaitGroup *sync.WaitGroup) {
	defer finishWaitGroup.Done()

	// columnCount is the amount of column in the CSV file. The value
	// "len(knownEditors)+1" is the mount of all editors plus column for
	// changeset count
	columnCount := len(keywords) + 1

	// First column contains the date, the others contain the data
	headLine := createHeadLine("date", keywords)

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

			comment := strings.ToLower(changeset.Comment)
			createdAt := changeset.CreatedAt[0:7] // e.g. "2020-04"

			if _, ok := aggregationMap[createdAt]; !ok {
				sigolo.Info("Create new map for '%s'", createdAt)
				aggregationMap[createdAt] = make(map[string]int)
			}

			for _, keyword := range keywords {
				if strings.Contains(comment, keyword) {
					aggregationMap[createdAt][keyword]++
				}
			}
		}

		sigolo.Info("Filtered %d changeset comments which took %dms", common.CACHE_SIZE, time.Since(clock).Milliseconds())
	}

	writeToFileWithDates(columnCount, keywords, false, aggregationMap, writer)
}
