package common

type Changeset struct {
	Id        int64
	User      string
	CreatedBy string
	CreatedAt string
}

// This defines the known editors later used for analysis of the changsets
var (
	CACHE_SIZE = 1000000

	EditorKey = "created_by" // Key in the tags

	UnknownEditor = "_UNKNOWN"
	NoEditor      = "_NO_EDITOR"
	KnownEditors  = []string{
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
		UnknownEditor,
		NoEditor,
	}
)
