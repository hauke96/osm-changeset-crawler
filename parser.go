// This file contains the parser creating changeset object from strings received
// by a given channel.
package main

import (
	"encoding/xml"
	"time"

	"github.com/hauke96/sigolo"
)

func parse(cacheSize int, changesetStringChannel <-chan []string, changesetChannel chan<- []Changeset) {
	defer close(changesetChannel)
	clock := time.Now()

	// Amount of processed changeset within the current cache. When the cache
	// is sent to the channel, this variable will be reset
	processedChangesets := 0
	cache := make([]Changeset, cacheSize)

	for changesets := range changesetStringChannel {
		sigolo.Info("Received %d changesets -> parsing", len(changesets))
		for i, changesetString := range changesets {
			// No data, no action
			if changesetString == "" {
				continue
			}

			var changeset Changeset
			xml.Unmarshal([]byte(changesetString), &changeset)

			cache[processedChangesets] = changeset
			processedChangesets++

			sigolo.Debug("Parsed and cached changeset with ID %d:", changeset.Id)
			sigolo.Debug("    Receive index : %d", i)
			sigolo.Debug("    Cache index   : %d", processedChangesets)

			if processedChangesets%cacheSize == 0 {
				sigolo.Info("Parsed %d changesets", cacheSize)
				sigolo.Info("  Parsing took %dms", time.Since(clock).Milliseconds())

				changesetChannel <- cache
				cache = make([]Changeset, cacheSize)
				processedChangesets = 0

				clock = time.Now()
			}
		}
	}

	// When there're actually remaining changesets, send them
	if processedChangesets != 0 {
		changesetChannel <- cache
	}
}
