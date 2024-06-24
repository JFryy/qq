# qq

`qq` is a interoperable configuration format transcoder with `jq` querying ability powered by gojq. `qq` is multi modal, and can be used as a replacement for `jq` or be interacted with via a repl with autocomplete and realtime rendering preview for building queries.

## Usage
Basic usage:
<a href="https://asciinema.org/a/665317" target="_blank"><img src="https://asciinema.org/a/665317.svg" /></a>

```sh
# query file and infer format from extension
qq .foo.bar file.xml

# query file through pipe, transcode xml -> terraform (it can be done but it probably shouldn't)
cat file.xml | qq '.bar.foo[].meep' -i xml -o tf

# interactive query builder mode on target file
qq . file.toml --interactive
```

## Installation

From source (requires `go` `>=1.22.4`)
```shell
make install
```

Download at releases [here](https://github.com/JFryy/qq/releases).

## Background

`qq` is inspired by `fq` and `jq`. `jq` is a powerful and succinct query tool, sometimes I would find myself needing to use another bespoke tool for another format than json, whether its something dedicated with json query built in or a simple converter from one configuration format to json to pipe into jq. `qq` aims to be the only utility needed for most interaction with structured formats in the terminal. It can transcode configuration formats interchangeably between one-another with the power of `jq` and it has an `an interactive repl (with automcomplete)` to boot so you can have an interactive experience when building queries optionally. Many thanks to the authors of the libraries used in this project, especially `jq`, `gojq`, and `fq` for direct usage or inspiration for the project.


## Features
* support a wide range of configuration formats and transform them interchangeably between eachother.
* quick and comprehensive querying of configuration formats without needing a pipeline of dedicated tools.
* provide a fun to use interactive mode for building queries with autocomplete and realtime rendering preview.
* `qq` is broad, but focuses on performance of encodings (but mostly `gojq` is very fast), execution is often times faster than most any "jq but for `${x}` configuration format"-type tools. `qq` performs similarly to benchmarks of `jq` running on `JSON` itself in most covered formats.


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
| HTML        | ❌ Not Supported | ❌ Not Supported |


## Caveats
* `qq` is not a full `jq`/`*q` replacement and comes with idiosyncrasies from the underlying `gojq` library.
* the encoders and decoders are not perfect and may not be able to handle all edge cases.


## Contributions
All contributions are welcome to `qq`, especially for upkeep/optimization/addition of new encodings. For ideas on contributions [please refer to the todo docs](https://github.com/JFryy/qq/blob/main/docs/TODO.md) or make an issue/PR for a suggestion if there's something that's wanted or fixes.

## Thanks and Acknowledgements / Related Projects
* [gojq](https://github.com/itchyny/gojq): `gojq` is a pure Go implementation of jq. It is used to power the query engine of qq.
* [fq](https://github.com/wader/fq) : fq is a `jq` like tool for querying a wide array of binary formats.
* Many encoding modules 🍻
