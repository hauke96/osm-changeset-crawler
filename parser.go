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

	// TODO parameter
	amountOfCunks := 5

	// Amount of processed changeset within the current cache. When the cache
	// is sent to the channel, this variable will be reset
	cacheIndex := 0
	cache := make([]Changeset, CACHE_SIZE)

	receivedChangesetSets := 0

	for changesets := range changesetStringChannel {
		sigolo.Info("Received changesets set %d -> parsing", receivedChangesetSets)
		receivedChangesetSets++

		finishChan := make(chan bool)
		chunkSize := len(changesets) / amountOfCunks
		for chunk := 0; chunk < amountOfCunks; chunk++ {
			startIndex := chunk * chunkSize
			go parseChangesets(&cache, startIndex, changesets[startIndex:startIndex+chunkSize], finishChan)
		}
		for chunk := 0; chunk < amountOfCunks; chunk++ {
			<-finishChan
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

func parseChangesets(cache *[]Changeset, cacheIndex int, changesets []string, finishChan chan bool) {
	name := cacheIndex
	sigolo.Info("%d start (at %d)", name, cacheIndex)
	for i, changesetString := range changesets {
		// No data, no action
		if changesetString == "" {
			continue
		}

		var changeset Changeset
		xml.Unmarshal([]byte(changesetString), &changeset)

		(*cache)[cacheIndex] = changeset
		cacheIndex++

		sigolo.Debug("Parsed and cached changeset with ID %d:", changeset.Id)
		sigolo.Debug("    Receive index : %d", i)
		sigolo.Debug("    Cache index   : %d", cacheIndex)
	}
	sigolo.Info("%d end (at %d)", name, cacheIndex-1)

	finishChan <- true
}
