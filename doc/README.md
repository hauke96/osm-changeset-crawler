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

A **analyser** is basically a function reading the incoming data and writes some result to a CSV file.
Currently there is no interface or higher abstraction but some functions shared between several analysers are used.


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

Regarding unmarshalling/deserializing of XML strings, I tested the following things/libraries:
1. `encoding/xml` unmarshal function: Very very slow and memory exhaustive when using `xml.Unmarshal(...)`
2. `encoding/xml` decoder and tokenizer: A bit faster but still very slow
3. Libraries using libxml: Faster but still quite slow. I tested [alexrsagen/go-libxml](https://github.com/alexrsagen/go-libxml) and [libxml2/parser](https://github.com/lestrrat-go/libxml2/parser).
4. Manual parsing (currently used): Fastest because we have knowledge about the XML structure and can use this.
