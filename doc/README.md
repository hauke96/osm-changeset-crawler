This is a file with some documentation for developers.

# Pipeline

.xml/.osm file → Reader → Parser → Analysers → CSV files

The **reader** (`reader.go`) reads the file line by line and searches for changesets.
This is simply done by searching for `<changeset ` and `</changeset>` (ans `/>` for single line changesets).
After a certain amount of changesets (specified by `CACHE_SIZE` from the `common.go` file), it's send via a channel to the parser.

The **parser** (`parser.go`) takes `CACHE_SIZE` many changeset strings (including tags and everything) and reads these strings to find the actual metadata.
To increase performance, the XML parsing is done manually with knowledge about the XML structure (things like: a tag only has `k` followed by `v` as attributes).
All the collected metadata is send to the analysers using a list of channels.
Each analyser has its own channel for the data.

An **analyser** (from `./analysis/analyser-*.go`) is basically a function reading the incoming data and writes some result to a CSV file.
Currently there is no interface or higher abstraction but some functions are shared between several analysers.


```
read    parse   analyser
  |        |        |
 _|_       |        |
|0  |      |        |
|   |      |        |
|___|     _|_       |
  |----->|0  |      |
 _|_     |||||      |
|1  |    |||||      |
|   |    |||||      |
|___|    |___|     _|_
  |        |----->|0  |
  |       _|_     |   |
  |----->|1  |    |___|
 _|_     |||||      |
|2  |    |||||      |
|   |    |||||      |
|___|    |___|     _|_
  |        |----->|1  |
  |       _|_     |   |
  |----->|2  |    |___|
 _|_     |||||      |
|3  |    |||||      |

  .        .        .
  .        .        .
  .        .        .

|n  |      |----->|n-1|
|___|     _|_     |   |
  |----->|n  |    |___|
  |      |||||      |
  X      |||||      |
         |||||      |
         |___|     _|_
           |----->|n  |
           |      |   |
           X      |___|
                    |
                    |
                result.csv

```

Parsing is the most complex task and uses the most time, therefore, it internally cuts the `CACHE_SIZE` many changeset strings further into chunks and parallelizes their processing.
When looking at the performance measurements (from the general [README.md](../README.md)) one can see, that the amount of analysers doesn't hav a significant impact on the overall performance.
That's because the reading and parsing takes so long.

Regarding performant unmarshalling/deserializing of XML strings into go objects, I tested the following things/libraries:
1. `encoding/xml` unmarshal function: Very very slow and memory exhaustive when using `xml.Unmarshal(...)`
2. `encoding/xml` decoder and tokenizer: A bit faster but still very slow
3. Libraries using libxml: Faster but still quite slow. I tested [alexrsagen/go-libxml](https://github.com/alexrsagen/go-libxml) and [libxml2/parser](https://github.com/lestrrat-go/libxml2/parser).
4. Manual parsing (currently used): Fastest because we have knowledge about the XML structure and can use this.

# Writing your own analyser

If you are familiar with go, this is a simple task.

1. Take a look at the existing analysers and try to understand what they do (not in detail but it's important to understand where the data comes from and where it's written)
2. Create a new analyser file like `./analysis/analyser-my-super-analyser.go`
3. Add a function with the following signature:
```go
func AnalyseMySuperAnalyser(outputPath string,
    changsetChannel <-chan []common.Changeset,
    finishWaitGroup *sync.WaitGroup) {
    ...
}
```
4. Setting up your analyser and the CSV file. Currently -- as there are not many abstractions -- building the CSV file depends heavily on the known editors.
```go
func ... {
    // The main file uses the wait groups to determine finished analysers
    defer finishWaitGroup.Done()

    // important for the csv file writer to know. The +1 is for the "_ALL" column which contains the sum of all values
	columnCount := len(common.KNOWN_EDITORS) + 1

    // The head line of the CSV file
	headLine := createEditorHeadLine(columnCount)

    // clock: To measure time and performance
    // aggregationMap: A map from user/month/... to a map from editor to number.
    // writer: The CSV file writer
	clock, aggregationMap, writer := initAnalyser(outputPath, headLine)

    // ==========
    // Your analysis goes here
    // ==========

    // Well, write the result to the file
	writeToFile(columnCount, aggregationMap, writer)
}
```
5. Write your analysis. Usually we want to iterate over all changesets, so this example does this as well:
```go
    // Init stuff is done here

    for changesets := range changsetChannel {
        for _, changeset := range changesets {
            // ID 0 indicates an empty place in the cache, where no changeset is. Happens when the last cache chunk is not completely filled.
            // Do some further filtering here
			if changeset.Id == 0 {
				continue
			}

            // Initialize maps for new keys
            if _, ok := aggregationMap[key]; !ok {
				aggregationMap[key] = make(map[string]int)
			}

            // Collect your information here, whatever it is
            ...

            // The "key" represents a line in the CSV file, where the value of "key" will appear in the first column of each line. The "col" variable is the actual column in each line.
			aggregationMap[key][col]++
        }
    }

    // Writing is done here
```
6. Add your analyser to the CLI in `main.go` and make is usable. The main-file parses the CLI arguments and flags using the `kingpin` library. A switch statement goes through all analysers passed to the application:
```go
    switch analyserString {
    case "my-super-analyser":
        superAnalyserChannel := make(chan []common.Changeset, 5)
        changesetChannels = append(changesetChannels, superAnalyserChannel)
        go analysis.AnalyseMySuperAnalyser("result_super-analyser.csv", superAnalyserChannel, &finishWaitGroup)
        finishWaitGroup.Add(1)
    ...
    }
```

Basically that's it.
You may want to add some logging and if you want to analyse things that are not based on editors, you probably have to write new function to init the analyser.

# Add missing tags to the parser
The parser only supports a fixed set of tags and attributes due to the manual XML parsing.
I do it manually, because it is much faster than e.g. the golang `encoding/xml` package.

The parser has an `unmarshal(...)` function going through a changeset string and trying to find tags.

The XML structure has attributes on a changeset like the user, chreated_at and so on.
However there're also true XML tags like `<tag k="foo" v="bar" />` which also hold valuable information.
To make things simple, everything (tags and attributed) are stored on the `common.Changeset` object next to each other as normal members of the object.

```go
switch k {
case "k": // <tag k="..." v="..."/>
    k = v // because we think of "comment=foo" and "comment" is the value here but we think of it as the actual key of the tag

    i++ // skip space between "k" and "v" XML elements
    i, _, v = readTag(i, data) // reads "v=..."

    switch k {
    case "created_by":
        c.CreatedBy = v
    case "comment":
        c.Comment = v
    case "source":
        c.Source = v
    //
    // --> Add new tag here
    //
    }
case "id":
    n, err := strconv.Atoi(v)
    sigolo.FatalCheck(err)
    c.Id = int64(n)
case "user":
    c.User = v
case "created_at":
    c.CreatedAt = v
//
// --> Add new attribute here
//
case "":
    break
default:
}
```

I marked the placed where new things should be added.
Just look at the existing cases and add your new things there.
Because you also need a new member in the `common.Changeset` object, you have to edit this too.
After that, you can use the new tag value in your analysis.
