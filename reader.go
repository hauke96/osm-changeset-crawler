package main

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/hauke96/sigolo"
)

// read the given file and cache "cacheSize" many changesets from that file
// before handing it over to the "pipeline". The pipeline receives an array of
// strings, each string is one changeset.
func read(fileName string, cacheSize int, pipeline chan []string) {
	changesetPrefix := "<ch"
	changesetSuffix := "</ch"
	changesetOneLineSuffix := "/>"

	cache := make([]string, cacheSize)

	// Open file
	fileHandle, err := os.Open(fileName)
	sigolo.FatalCheck(err)
	defer fileHandle.Close()
	sigolo.Info("Opened file")

	var line string
	processedChangesets := 0
	scanner := bufio.NewScanner(fileHandle)
	sigolo.Info("Created scanner")
	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())

		// New changeset starts
		if strings.HasPrefix(line, changesetPrefix) {
			sigolo.Debug("Start of changeset")
			sigolo.Debug("  %s", line)
			// Read all lines of this changeset
			changesetString := line

			// If the read line is not a one-line-changeset like
			// "<changeset id=123 open=false ... />"), then read the other lines
			if !strings.HasSuffix(changesetString, changesetOneLineSuffix) {
				for scanner.Scan() {
					line = strings.TrimSpace(scanner.Text())
					sigolo.Debug("    %s", line)
					changesetString += line

					// Changeset ends
					if strings.HasPrefix(line, changesetSuffix) {
						sigolo.Debug("End of changeset")
						break
					}
				}
			}

			// Done reading the changeset, add it to the cache
			cache[processedChangesets] = changesetString
			processedChangesets++

			sigolo.Debug("=> %s", changesetString)
		}

		if processedChangesets > 0 && processedChangesets%cacheSize == 0 {
			sigolo.Info("Handled %d changesets", processedChangesets)

			pipeline <- cache

			processedChangesets = 0
			cache = make([]string, cacheSize)

			time.Sleep(1 * time.Second)
		}
	}

	pipeline <- cache
	close(pipeline)
}
