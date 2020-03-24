# OSM changeset analyser
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
Here a short version of the `--help` flag:
```
usage: OSM changeset analyser --analysers=ANALYSERS [<flags>] <file>

A tool analysing the changesets from OpenStreetMap (OSM).

Flags:
  -h, --help                 Show context-sensitive help (also try --help-long and --help-man).
  -d, --debug                Verbose mode, showing additional debug information
      --analysers=ANALYSERS  A comma separated list of analysers
  -v, --version              Show application version.

Args:
  <file>  The file to analyse

ANALYSERS:
  The 'analysers' flag is a comma separated list of analysers all creating their own CSV file:

  * editor-count : Counts the amount of the most common editors for each month.
  * no-source-count : Counts the amount of monthly changesets without source tag, sorted by editor.
  * user-without-source : Counts for each user the amount of changesets without source tag for each editor editor.
  * comment-keywords(foo,bar) : Takes keywords (in this case "foo" and "bar") and counts their occurrence per month. Comments and keywords are converted into lower case.
```

So for example this call analyses the `data.osm` using the three analysers for the editor count, the editor without source and the users without source:
```bash
$> go build .
$> ./osm-changeset-analyser --analysers=editor-count,no-source-count,user-without-source data.osm
$> ll result*
-rw-r--r-- 1 hauke hauke 8,2K  7. Mär 15:03 result_editor-count.csv
-rw-r--r-- 1 hauke hauke 8,2K  7. Mär 15:03 result_no-source-count.csv
-rw-r--r-- 1 hauke hauke  529  7. Mär 15:03 result_user-without-source.csv
```

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

## Performance
I tested the performance on my private computer (s. below).
Of course there were some other applications running (like E-Mail client, Browser, Editors, etc.) but I wasn't doing anything during the execution.

### Dataset
I used the [changesets-200224.osm.bz2](https://planet.openstreetmap.org/planet/2020/changesets-200224.osm.bz2) (donwload size: 3.2GB / decompressed size: 34GB).

### My system:
* CPU: Intel Xeon E3-1231 v3, 8x3.4GHz
* RAM: 16GB DDR3 1333MHz
* Drive: Samsung SSD 850 EVO

### Measurements

Here are some example executions:

| active analysers | execution time | processing speed | RAM usage (approx.) |
|:-- |:-- |:-- |:-- |
| `no-editor` | 6m, 39s | 85 MB/s | 6.8 GB |
| `user-without-source` | 7m, 12s | 78 MB/s | approx. 10 GB |
| `no-editor` <br> `no-source-count` <br> `user-without-source` | 7m, 21s | 77 MB/s | 10GB |

### Output files
```bash
13K result_editor-count.csv
13K result_no-source-count.csv
52M result_user-without-source.csv
```

# For developers
There exist multiple goroutines processing the data asynchronously.
See the [doc](doc/README.md) folder for more information.
