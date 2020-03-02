// This file contains the parser creating changeset object from strings received
// by a given channel.
package main

import (
	"encoding/xml"

	"github.com/hauke96/sigolo"
)

func parse(cacheSize int, changesetStringChannel <-chan []string, changesetChannel chan<- []Changeset) {
	defer close(changesetChannel)

	// Amount of processed changeset within the current cache. When the cache
	// is sent to the channel, this variable will be reset
	processedChangesets := 0
	cache := make([]Changeset, cacheSize)

	for changesets := range changesetStringChannel {
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
				sigolo.Info("Send parsed changesets")
				changesetChannel <- cache
				cache = make([]Changeset, cacheSize)
				processedChangesets = 0
			}
		}
	}

	// When there're actually remaining changesets, send them
	if processedChangesets != 0 {
		changesetChannel <- cache
	}
}
