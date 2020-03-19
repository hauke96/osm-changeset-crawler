package common

type Changeset struct {
	Id        int64
	User      string
	CreatedBy string
	CreatedAt string
	Comment   string
	Source    string
}

// This defines the known editors later used for analysis of the changsets
var (
	CACHE_SIZE = 1000000

	// BEWARE:
	// The order of this is important as "id" is also contained in "maps.me android" so that MAPS.ME would considered to be iD.
	KNOWN_EDITORS = []string{
		"maps.me",
		"potlatch",
		"osmand+",
		"vespucci",
		"streetcomplete",
		"osmtools",
		"merkaartor",
		"osm2go",
		"josm",
		"id",
		EDITOR_UNKNOWN,
		EDITOR_NOT_SET,
	}
)

const (
	KEY_CREATED_BY = "created_by" // Key in the tags

	ALL = "_ALL"

	EDITOR_UNKNOWN = "_UNKNOWN"
	EDITOR_NOT_SET = "_NO_EDITOR"

	USER_NOT_SET = "_NO_USER"
)
