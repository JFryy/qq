# qq

`qq` is a interoperable configuration format transcoder with `jq` querying ability powered by gojq. `qq` is multi modal, and can be used as a replacement for `jq` or be interacted with via a repl with autocomplete and realtime rendering preview for building queries.

## Usage
Basic usage:
<a href="https://asciinema.org/a/665317" target="_blank"><img src="https://asciinema.org/a/665317.svg" /></a>

```sh
# query file and infer format
qq . foo.bar file.xml

# query file with explicit format, transcode to another format
cat file.xml | qq . foo.bar -i xml -o hcl

# interactive query builder
qq . file.toml --interactive
```

## Installation

From source (requires `go` `>=1.22.4`)
```shell
make install
```

Download at releases [here](https://github.com/JFryy/qq/releases).

## Background

`qq` is heavily inspired by `fq` and `jq`. `jq` is a powerful and succinct query tool, sometimes I would find myself needing to use another bespoke tool for another format than json, whether its something dedicated with json query built in or a simple converter from one configuration format to json to pipe into jq. `qq` aims for the lofty goal to be the only utility needed for majority of interaction with structured formats in the terminal. It combines transcoding configuration formats from one to another, the power of `jq`, and `an interactive repl (with automcomplete)` for building more advanced queries. Many thanks to the authors of the libraries used in this project, especially the `jq` and `gojq` authors and the authors of the many encoding libraries used in this project.


## Goals
* support a wide range of configuration formats and transform them interchangeably between eachother.
* quick and comprehensive querying of configuration formats without needing a pipeline of dedicated tools.
* provide a fun to use interactive mode for building queries with autocomplete and realtime rendering preview.


## Supported formats
| Format      | Input          | Output         |
|-------------|----------------|----------------|
| JSON        | ✅ Supported   | ✅ Supported   |
| YAML        | ✅ Supported   | ✅ Supported   |
| TOML        | ✅ Supported   | ✅ Supported   |
| XML         | ✅ Supported   | ✅ Supported   |
| INI         | ✅ Supported   | ✅ Supported   |
| HCL         | ✅ Supported   | ✅ Supported   |
| TF          | ✅ Supported   | ✅ Supported   |
| CSV         | ✅ Supported   | ❌ Not Supported |
| Protobuf    | ❌ Not Supported | ❌ Not Supported |


## Caveats
* `qq` is not a full `jq`/`*q` replacement and comes with idiosyncrasies from the underlying `gojq` library.
* the encoders and decoders are not perfect and may not be able to handle all edge cases.

## Thanks and Acknowledgements / Related Projects
* [gojq](https://github.com/itchyny/gojq): `gojq` is a pure Go implementation of jq. It is used to power the query engine of qq.
* [fq](https://github.com/wader/fq) : fq is a `jq` like tool for querying a wide array of binary formats.

