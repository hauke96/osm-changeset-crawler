package main

import (
	"sync"

	"github.com/hauke96/osm-changeset-analyser/analysis"
	"github.com/hauke96/osm-changeset-analyser/common"

	"github.com/hauke96/sigolo"
)

func main() {
	//sigolo.LogLevel = sigolo.LOG_PLAIN
	//sigolo.LogLevel = sigolo.LOG_DEBUG

	changesetStringChan := make(chan []string, 5)
	changesetChannels := make([]chan<- []common.Changeset, 0)

	// editorCountChannel := make(chan []common.Changeset, 5)
	// noCommentCountChannel := make(chan []common.Changeset, 5)
	userWithoutSourceChannel := make(chan []common.Changeset, 5)

	// changesetChannels = append(changesetChannels, editorCountChannel)
	// changesetChannels = append(changesetChannels, noCommentCountChannel)
	changesetChannels = append(changesetChannels, userWithoutSourceChannel)

	finishWaitGroup := sync.WaitGroup{}
	finishWaitGroup.Add(3)

	// go read("/home/hauke/Dokumente/OSM/changeset-analysis/test.osm", changesetStringChan, &finishWaitGroup)
	go read("/home/hauke/Dokumente/OSM/changeset-analysis/changesets-200224.osm", changesetStringChan, &finishWaitGroup)
	// go read("test.osm", changesetStringChan, &finishWaitGroup)

	go parse(changesetStringChan, changesetChannels, &finishWaitGroup)

	// go analysis.AnalyseEditorCount("result_editor-count.csv", editorCountChannel, &finishWaitGroup)
	// go analysis.AnalyseNoSourceCount("result_no-source-count.csv", noCommentCountChannel, &finishWaitGroup)
	go analysis.AnalyseUserWithoutSource("result_user-without-source.csv", userWithoutSourceChannel, &finishWaitGroup)

	finishWaitGroup.Wait()

	sigolo.Info("Done")
}
