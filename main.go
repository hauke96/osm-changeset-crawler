package main

import (
	"bufio"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	// "time"

	"github.com/hauke96/sigolo"
)

type Osm struct {
	Changesets []Changeset `xml:"changeset"`
}

type Changeset struct {
	Id   int64  `xml:"id,attr"`
	User string `xml:"user,attr"`
	Tags []Tag  `xml:"tag"`
}

type Tag struct {
	K string `xml:"k,attr"`
	V string `xml:"v,attr"`
}

var (
	unknownEditor = "_UNKNOWN"
	noEditor      = "_NO_EDITOR"
	knownEditors  = []string{
		"josm",
		"id",
		"potlatch",
		"maps.me",
		"osmand+",
		"vespucci",
		"streetcomplete",
		"osmtools",
		"merkaartor",
		"osm2go",
		unknownEditor,
		noEditor,
	}
)

func main() {
	// All this is old and will be replaces soon:

	//sigolo.LogLevel = sigolo.LOG_PLAIN
	//sigolo.LogLevel = sigolo.LOG_DEBUG

	// time := "2020-02-28T12:00:00Z" // (time.Now()).Format(time.RFC3339)
	// bbox := "6,47,15,55"
	// bbox := "9.65,53.43,10.3,53.7"
	// osm := downloadChangesets(bbox, time)

	//osm := readChangesets("/home/hauke/Dokumente/OSM/changeset-analysis/test.osm", changesetFilter)
	// osm := readChangesets("/media/hauke/Dokumente/OpenStreetMap/changesets-200224.osm", changesetFilter)
	// osm := readChangesets("/home/hauke/Dokumente/OSM/changeset-analysis/changesets-200224.osm", changesetFilter)

	// sigolo.Info("Got %d changesets", len(osm.Changesets))

	// editorMap := sortByEditor(osm.Changesets)

	// sigolo.Info("JOSM: %d", len(editorMap["josm"]))
}

func readChangesets(fileName string, filter func(cs *Changeset) bool) *Osm {
	result := Osm{}
	result.Changesets = make([]Changeset, 0)

	changesetPrefix := "<ch"
	changesetSuffix := "</ch"
	changesetOneLineSuffix := "/>"

	cache := InitArrayCache(100000)

	sigolo.Debug("Read file '%s'", fileName)

	// Open file
	fileHandle, err := os.Open(fileName)
	sigolo.FatalCheck(err)
	defer fileHandle.Close()
	sigolo.Info("Opened file")

	// Read file line by line and find changesets within this file. Use the
	// cache to only keep a certain amount of changesets. The idea: Is the file
	// sorted by date, then the changesets in the end are the most recent ones.
	var line string
	processedChangesets := 0
	scanner := bufio.NewScanner(fileHandle)
	sigolo.Info("Created scanner")
	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())

		// New changeset starts
		if strings.HasPrefix(line, changesetPrefix) {
			sigolo.Debug("Start of changeset")
			sigolo.Debug("  %s", line)
			// Read all lines of this changeset
			changesetString := line

			// If the read line is not a one-line-changeset like
			// "<changeset id=123 open=false ... />"), then read the other lines
			if !strings.HasSuffix(changesetString, changesetOneLineSuffix) {
				for scanner.Scan() {
					line = strings.TrimSpace(scanner.Text())
					sigolo.Debug("    %s", line)
					changesetString += line

					// Changeset ends
					if strings.HasPrefix(line, changesetSuffix) {
						sigolo.Debug("End of changeset")
						break
					}
				}
			}
			cache.AddElement(&changesetString)

			sigolo.Debug("=> %s", changesetString)
		}

		processedChangesets++
		if processedChangesets%cache.MaxSize == 0 {
			changesets := parseChangesetStrings(cache.Elements)
			sigolo.Info("Handled %d changesets", processedChangesets)

			writeEditorCount(processedChangesets, changesets)
		}
	}

	err = scanner.Err()
	sigolo.FatalCheck(err)

	changesets := parseChangesetStrings(cache.Elements)

	writeEditorCount(processedChangesets, changesets)

	result.Changesets = changesets
	return &result // parseOsm(fileHandle)
}

func writeEditorCount(processedChangesets int, changesets []Changeset) {
	changesetMap := sortByEditor(changesets)

	file, err := os.OpenFile("result.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	sigolo.FatalCheck(err)
	defer file.Close()

	stat, err := os.Stat("result.csv")
	firstWrite := stat.Size() == 0

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// If no frist line exist, write the head line
	if firstWrite {
		sigolo.Info("Write head line to CSV file")

		headLine := make([]string, len(knownEditors)+1) // All editor plus column for changeset count
		headLine[0] = "changeset count"
		for i := 0; i < len(knownEditors); i++ {
			headLine[i+1] = knownEditors[i]
		}

		err = writer.Write(headLine)
		sigolo.FatalCheck(err)
	}

	line := make([]string, len(knownEditors)+1) // All editor plus column for changeset count
	line[0] = strconv.Itoa(processedChangesets)

	sigolo.Info("Result for %d:", processedChangesets)
	for i := 0; i < len(knownEditors); i++ {
		editor := knownEditors[i]
		editorCount := len(changesetMap[editor])

		line[i+1] = strconv.Itoa(editorCount)

		sigolo.Info("  %s : %d", editor, editorCount)
	}

	err = writer.Write(line)
	sigolo.FatalCheck(err)
}

func downloadChangesets(bbox, time string) *Osm {
	sigolo.Debug("Params:")
	sigolo.Debug("  time = %s", time)
	sigolo.Debug("  bbox = %s", bbox)

	urlString := "https://www.openstreetmap.org/api/0.6/changesets?bbox=" + bbox + "&time=" + time
	sigolo.Debug("GET %s", urlString)

	httpResponse, err := http.Get(urlString)
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != 200 {
		err = errors.New(httpResponse.Status)
	}
	sigolo.FatalCheck(err)
	sigolo.Debug(httpResponse.Status)

	return parseOsm(httpResponse.Body)
}

func parseOsm(stream io.Reader) *Osm {
	xmlBytes, err := ioutil.ReadAll(stream)
	sigolo.FatalCheck(err)
	xmlString := string(xmlBytes)
	sigolo.Plain(xmlString)

	var osm Osm
	err = xml.Unmarshal(xmlBytes, &osm)
	sigolo.FatalCheck(err)
	return &osm
}

func parseChangesetStrings(changesetStrings []*string) []Changeset {
	changesets := make([]Changeset, 0)

	for i, changesetString := range changesetStrings {
		// If the cache is not fully filled
		if changesetString == nil {
			continue
		}

		var changeset Changeset
		xml.Unmarshal([]byte(*changesetString), &changeset)
		changesets = append(changesets, changeset)
		sigolo.Debug("%d => %v", i, changeset)
	}

	return changesets
}

func sortByEditor(changesets []Changeset) map[string][]Changeset {
	result := make(map[string][]Changeset)

	for _, changeset := range changesets {
		var createdByTag string

		// Get "created_by" tag from changeset
		tags := changeset.Tags
		for _, tag := range tags {
			sigolo.Debug("Check tag key '%s' with value '%v'", tag.K, tag.V)
			if tag.K == "created_by" {
				createdByTag = strings.ToLower(tag.V)
			}
		}

		if createdByTag == "" {
			sigolo.Debug("No editor found for changeset %d", changeset.Id)
			createdByTag = noEditor
			continue
		}

		// Check if the "created_by" value is known. The value can be
		// complicated like "JOSM/1.5 (15628 en_GB)" so we have to do some minor
		// parsing here
		knownEditorFound := false
		for _, knownEditor := range knownEditors {
			if strings.Contains(createdByTag, knownEditor) {
				createdByTag = knownEditor
				knownEditorFound = true
			}
		}

		if !knownEditorFound {
			sigolo.Debug("Unknown editor found: %s", createdByTag)
			createdByTag = unknownEditor
			continue
		}

		sigolo.Debug("Changeset '%d': Editor '%s' found", changeset.Id, createdByTag)

		// Add CS to map
		if result[createdByTag] == nil {
			result[createdByTag] = make([]Changeset, 1)
		}
		result[createdByTag] = append(result[createdByTag], changeset)
	}

	return result
}

func changesetFilter(cs *Changeset) bool {
	return true
}
