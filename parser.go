// This file contains the parser creating changeset object from strings received
// by a given channel.
package main

import (
	"encoding/xml"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/hauke96/sigolo"
)

func parse(changesetStringChannel <-chan []string, changesetChannel chan<- []Changeset) {
	defer close(changesetChannel)
	clock := time.Now()

	// TODO parameter
	amountOfCunks := int(math.Min(10, float64(CACHE_SIZE)))

	// Amount of processed changeset within the current cache. When the cache
	// is sent to the channel, this variable will be reset
	cacheIndex := 0
	cache := make([]Changeset, CACHE_SIZE)

	receivedChangesetSets := 0

	for changesets := range changesetStringChannel {
		sigolo.Info("Received changesets set %d -> parsing", receivedChangesetSets)
		sigolo.Info("  Receiving took %dms", time.Since(clock).Milliseconds())
		clock = time.Now()
		receivedChangesetSets++

		// Parallelize parsing
		finishChan := make(chan bool)
		chunkSize := len(changesets) / amountOfCunks
		sigolo.Debug("Chunk size: %d", chunkSize)
		for chunk := 0; chunk < amountOfCunks; chunk++ {
			startIndex := chunk * chunkSize
			go parseChangesets(&cache, startIndex, changesets[startIndex:startIndex+chunkSize], finishChan)
		}
		for chunk := 0; chunk < amountOfCunks; chunk++ {
			<-finishChan
		}

		sigolo.Info("Parsed %d changesets", CACHE_SIZE)
		sigolo.Info("  Parsing took %dms", time.Since(clock).Milliseconds())
		clock = time.Now()

		changesetChannel <- cache
		cache = make([]Changeset, CACHE_SIZE)
		cacheIndex = 0
		sigolo.Info("  Sending parsed data took %dms", time.Since(clock).Milliseconds())
		clock = time.Now()
	}

	// When there're actually remaining changesets, send them
	if cacheIndex != 0 {
		changesetChannel <- cache
	}
}

func parseChangesets(cache *[]Changeset, cacheIndex int, changesets []string, finishChan chan bool) {
	for i, changesetString := range changesets {
		// No data, no action
		if changesetString == "" {
			continue
		}

		// var changeset Changeset
		// xml.Unmarshal([]byte(changesetString), &changeset)

		changeset := unmarshal(changesetString)
		(*cache)[cacheIndex] = changeset
		cacheIndex++

		sigolo.Debug("Parsed and cached changeset with ID %d:", changeset.Id)
		sigolo.Debug("    Receive index : %d", i)
		sigolo.Debug("    Cache index   : %d", cacheIndex)
	}

	finishChan <- true
}

func unmarshal(changeset string) Changeset {
	decoder := xml.NewDecoder(strings.NewReader(changeset))

	c := Changeset{}

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		sigolo.FatalCheck(err)

		switch root := tok.(type) {
		case xml.StartElement:
			switch root.Name.Local {
			case "changeset":
				for _, a := range root.Attr {
					switch a.Name.Local {
					case "id":
						i, err := strconv.Atoi(a.Value)
						sigolo.FatalCheck(err)
						c.Id = int64(i)
						break
					case editorKey:
						c.User = a.Value
						break
					}
				}
			case "tag":
				t := Tag{}
				for _, a := range root.Attr {
					switch a.Name.Local {
					case "k":
						t.K = a.Value
						break
					case "v":
						t.V = a.Value
						break
					}
				}
				c.Tags = append(c.Tags, t)
			}
		}
	}

	return c
}
