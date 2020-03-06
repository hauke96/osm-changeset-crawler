package main

import (
	"github.com/hauke96/osm-changeset-analyser/analysis"
	"github.com/hauke96/osm-changeset-analyser/common"

	"github.com/hauke96/sigolo"
)

func main() {
	//sigolo.LogLevel = sigolo.LOG_PLAIN
	//sigolo.LogLevel = sigolo.LOG_DEBUG

	changesetStringChan := make(chan []string, 5)
	changesetChannels := make([]chan<- []common.Changeset, 1)

	editorCountChannel := make(chan []common.Changeset, 5)

	changesetChannels = append(changesetChannels, editorCountChannel)

	// go read("/home/hauke/Dokumente/OSM/changeset-analysis/test.osm", changesetStringChan)
	go read("/home/hauke/Dokumente/OSM/changeset-analysis/changesets-200224.osm", changesetStringChan)
	// go read("test.osm", changesetStringChan)

	go parse(changesetStringChan, changesetChannels)

	analysis.Analyse("result.csv", editorCountChannel)

	sigolo.Info("Done")
}
