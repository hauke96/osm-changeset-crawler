// This file contains the parser creating changeset object from strings received
// by a given channel.
package main

import (
	"math"
	"strconv"
	"time"

	"github.com/hauke96/osm-changeset-analyser/common"
	"github.com/hauke96/sigolo"
)

func parse(changesetStringChannel <-chan []string, changesetChannel []chan<- []common.Changeset) {
	defer func() {
		for _, c := range changesetChannel {
			close(c)
		}
	}()

	clock := time.Now()

	// TODO parameter
	amountOfCunks := int(math.Min(20, float64(common.CACHE_SIZE)))

	// Amount of processed changeset within the current cache. When the cache
	// is sent to the channel, this variable will be reset
	cacheIndex := 0
	cache := make([]common.Changeset, common.CACHE_SIZE)

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

		sigolo.Info("Parsed %d changesets", common.CACHE_SIZE)
		sigolo.Info("  Parsing took %dms", time.Since(clock).Milliseconds())
		clock = time.Now()

		for _, c := range changesetChannel {
			c <- cache
		}

		cache = make([]common.Changeset, common.CACHE_SIZE)
		cacheIndex = 0
		sigolo.Info("  Sending parsed data took %dms", time.Since(clock).Milliseconds())
		clock = time.Now()
	}

	// When there're actually remaining changesets, send them
	if cacheIndex != 0 {
		for _, c := range changesetChannel {
			c <- cache
		}
	}
}

func parseChangesets(cache *[]common.Changeset, cacheIndex int, changesets []string, finishChan chan bool) {
	for _, changesetString := range changesets {
		// No data, no action
		if changesetString == "" {
			continue
		}

		// var changeset Changeset
		// xml.Unmarshal([]byte(changesetString), &changeset)

		changeset := unmarshal(changesetString)
		(*cache)[cacheIndex] = changeset
		cacheIndex++

		// sigolo.Debug("Parsed and cached changeset with ID %d:", changeset.Id)
		// sigolo.Debug("    Receive index : %d", i)
		// sigolo.Debug("    Cache index   : %d", cacheIndex)
	}

	finishChan <- true
}

func unmarshal(data string) common.Changeset {
	c := common.Changeset{}

	i := 11 // skip the beginning of "<changeset "
	l := len(data)
	var k, v string

	for i < l {
		if data[i] == ' ' || data[i] == '/' || data[i] == '<' || data[i] == '>' {
			i++
			continue
		}

		i, k, v = readTag(i, data)
		// sigolo.Debug("Found k='%s' and v='%s'", k, v)

		switch k {
		case "k": // <tag k="..." v="..."/>
			k = v

			i++ // skip space between "k" and "v" XML elements
			i, _, v = readTag(i, data)

			// sigolo.Debug("Found tag '%s'='%s'", k, v)

			switch k {
			case "created_by":
				c.CreatedBy = v
			}
		case "id":
			n, err := strconv.Atoi(v)
			sigolo.FatalCheck(err)
			c.Id = int64(n)
		case "user":
			c.User = v
		case "created_at":
			c.CreatedAt = v
		case "":
			break
		default:
			// sigolo.Debug("Unknown tag: '%s'='%s'", k, v)
		}

		k = ""
		v = ""
	}

	return c
}

// Parse something like: created_by="JOSM/2019"
// This will return (i, "created_by", "JOSM/2019")
func readTag(i int, data string) (int, string, string) {
	var k, v string

	for data[i] != '=' && data[i] != '>' && data[i] != ' ' {
		k += string(data[i])
		i++
	}

	if data[i] == ' ' { // we read the space after a XML tag beginning
		// sigolo.Debug("Found beginning of XML tag '%s'", k)
		i++
		return i, "", ""
	}

	if data[i] == '>' { // end of some tag
		// sigolo.Debug("Found ending of XML tag '%s'", k)
		i++
		return i, "", ""
	}

	i += 2 // skip ="

	for data[i] != '"' {
		v += string(data[i])
		i++
	}

	i++ // skip "

	return i, k, v
}
