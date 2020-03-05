package main

type Changeset struct {
	Id        int64
	User      string
	CreatedBy string
	CreatedAt string
}

// This defines the known editors later used for analysis of the changsets
var (
	editorKey = "created_by" // Key in the tags

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
