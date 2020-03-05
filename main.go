package main

import (
	"github.com/hauke96/sigolo"
)

var (
	CACHE_SIZE int
)

func main() {
	//sigolo.LogLevel = sigolo.LOG_PLAIN
	//sigolo.LogLevel = sigolo.LOG_DEBUG

	changesetStringChan := make(chan []string, 5)
	changesetChan := make(chan []Changeset, 5)

	CACHE_SIZE = 1000000

	// go read("/home/hauke/Dokumente/OSM/changeset-analysis/test.osm", changesetStringChan)
	go read("/home/hauke/Dokumente/OSM/changeset-analysis/changesets-200224.osm", changesetStringChan)
	// go read("test.osm", changesetStringChan)

	go parse(changesetStringChan, changesetChan)

	analyseEditorCount("result.csv", changesetChan)

	sigolo.Info("Done")
}
