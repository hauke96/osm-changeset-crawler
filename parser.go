// This file contains the parser creating changeset object from strings received
// by a given channel.
package main

import (
	"encoding/xml"
	"time"

	"github.com/hauke96/sigolo"
)

func parse(changesetStringChannel <-chan []string, changesetChannel chan<- []Changeset) {
	defer close(changesetChannel)
	clock := time.Now()

	// Amount of processed changeset within the current cache. When the cache
	// is sent to the channel, this variable will be reset
	cacheIndex := 0
	cache := make([]Changeset, CACHE_SIZE)

	receivedChangesetSets := 0

	for changesets := range changesetStringChannel {
		sigolo.Info("Received changesets set %d -> parsing", receivedChangesetSets)
		receivedChangesetSets++

		for i, changesetString := range changesets {
			// No data, no action
			if changesetString == "" {
				continue
			}

			var changeset Changeset
			xml.Unmarshal([]byte(changesetString), &changeset)

			cache[cacheIndex] = changeset
			cacheIndex++

			sigolo.Debug("Parsed and cached changeset with ID %d:", changeset.Id)
			sigolo.Debug("    Receive index : %d", i)
			sigolo.Debug("    Cache index   : %d", cacheIndex)
		}

		sigolo.Info("Parsed %d changesets", CACHE_SIZE)
		sigolo.Info("  Parsing took %dms", time.Since(clock).Milliseconds())

		changesetChannel <- cache
		cache = make([]Changeset, CACHE_SIZE)
		cacheIndex = 0

		clock = time.Now()
	}

	// When there're actually remaining changesets, send them
	if cacheIndex != 0 {
		changesetChannel <- cache
	}
}
