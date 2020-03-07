package main

import (
	"os"
	"strings"
	"sync"

	"github.com/hauke96/osm-changeset-analyser/analysis"
	"github.com/hauke96/osm-changeset-analyser/common"

	"github.com/hauke96/kingpin"
	"github.com/hauke96/sigolo"
)

const VERSION string = "v0.0.1"

var (
	app          = kingpin.New("OSM changeset analyser", "A tool analysing the changesets from OpenStreetMap (OSM).")
	appDebug     = app.Flag("debug", "Verbose mode, showing additional debug information").Short('d').Bool()
	appAnalysers = app.Flag("analysers", "A comma separated list of analysers").Required().String()
	appFile      = app.Arg("file", "The file to analyse").Required().String()
)

func configureCliArgs() {
	app.Author("Hauke Stieler")
	app.Version(VERSION)

	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')

	app.CustomDescription("ANALYSERS", `The 'analysers' flag is a comma separated list of analysers all creating their own CSV file:

* editor-count : Counts the amount of the most common editors for each month.
	Example output:
		changeset count,josm,id
		2020-01,12,30
		2020-02,15,39

* no-source-count : Counts the amount of monthly changesets without source tag, sorted by editor.
	Example output:
		changeset count,josm,id
		2020-01,0,5
		2020-02,2,10

* user-without-source : Counts for each user the amount of changesets without source tag for each editor editor.
	Example output:
		user,josm,id
		john,55,8
		anna,18,76
`)

}

func configureLogging() {
	if *appDebug {
		sigolo.LogLevel = sigolo.LOG_DEBUG
	} else {
		sigolo.LogLevel = sigolo.LOG_INFO
	}
}

func main() {
	configureCliArgs()
	_, err := app.Parse(os.Args[1:])
	sigolo.FatalCheck(err)
	configureLogging()

	changesetStringChan := make(chan []string, 5)
	changesetChannels := make([]chan<- []common.Changeset, 0)

	finishWaitGroup := sync.WaitGroup{}
	finishWaitGroup.Add(2) // reader and parser

	for _, analyserString := range strings.Split(*appAnalysers, ",") {
		switch analyserString {
		case "editor-count":
			editorCountChannel := make(chan []common.Changeset, 5)
			changesetChannels = append(changesetChannels, editorCountChannel)
			go analysis.AnalyseEditorCount("result_editor-count.csv", editorCountChannel, &finishWaitGroup)
			finishWaitGroup.Add(1)
		case "no-source-count":
			noSourceCountChannel := make(chan []common.Changeset, 5)
			changesetChannels = append(changesetChannels, noSourceCountChannel)
			go analysis.AnalyseNoSourceCount("result_no-source-count.csv", noSourceCountChannel, &finishWaitGroup)
			finishWaitGroup.Add(1)
		case "user-without-source":
			userWithoutSourceChannel := make(chan []common.Changeset, 5)
			changesetChannels = append(changesetChannels, userWithoutSourceChannel)
			finishWaitGroup.Add(1)
			go analysis.AnalyseUserWithoutSource("result_user-without-source.csv", userWithoutSourceChannel, &finishWaitGroup)
		}
	}

	go read(*appFile, changesetStringChan, &finishWaitGroup)
	go parse(changesetStringChan, changesetChannels, &finishWaitGroup)

	finishWaitGroup.Wait()

	sigolo.Info("Done")
}
