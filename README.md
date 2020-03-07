# osm-changeset-analyser
A tool analysing the changesets from [OpenStreetMap (OSM)](https://osm.org).

# Compilation
This uses [`sigolo`](https://github.com/hauke96/sigolo) (logging) and [`kingpin`](https://github.com/hauke96/kingpin) (CLI options) as dependencies.
Everything can be compiled normally.

```bash
go get https://github.com/hauke96/sigolo
go get https://github.com/hauke96/kingpin
go run .
```

# Usage
Currently no CLI options exist, this will come soon.
For now you have to edit the `main.go` file in order to change the behavior.

## Input data and format
OSM changesets have a simple XML structure. Each changeset has basic metadata (user, location, creation date, etc.) and more specific metadata (comment, source of data, etc.), which can consist of arbitrary XML tags.

```xml
<changeset id="1234567"
		created_at="2020-01-12T14:03:44Z"
		open="false"
		comments_count="2"
		changes_count="154"
		closed_at="020-01-12T14:04:15Z"
		min_lat="10.24"
		min_lon="20.48"
		max_lat="5.12"
		max_lon="2.56"
		uid="12345"
		user="mega-mapper-3000">
	<tag k="source" v="survey; Bing"/>
	<tag k="hashtags" v="#github;#example"/>
	<tag k="created_by" v="JOSM/1.5 (15492 en)"/>
	<tag k="comment" v="Useful information for other mappers"/>
</changeset>
```

The latest data for the whole planet can be downloaded from https://planet.openstreetmap.org/planet/changesets-latest.osm.bz2.
This is over 3GB large (decompressed approx. 34GB) and contains all changesets from 2005 til now.

# For developer
There exist multiple goroutines processing the data asynchronously.
See the [doc](doc/README.md) folder for more information.
