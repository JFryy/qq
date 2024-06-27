# qq

`qq` is a interoperable configuration format transcoder with `jq` query syntax powered by `gojq`. `qq` is multi modal, and can be used as a replacement for `jq` or be interacted with via a repl with autocomplete and realtime rendering preview for building queries.

## Usage
Basic usage:
<a href="https://asciinema.org/a/665317" target="_blank"><img src="https://asciinema.org/a/665317.svg" /></a>



```sh
# JSON is default in
cat file.${ext} | qq -i ${ext}

# query xml, grep with gron using qq io
qq file.xml -o gron | grep -vE "sweet.potatoes" | qq -i gron

# interactive query builder mode on target file
qq . file.toml --interactive
```

## Installation

From source (requires `go` `>=1.22.4`)
```shell
make install
```

Download at releases [here](https://github.com/JFryy/qq/releases).

Docker quickstart:

```shell
# install the image
docker pull jfryy/qq

# run an example
echo '{"foo":"bar"}' | docker run -i jfryy/qq '.foo = "bazz"' -o tf
```
## Background

`qq` is inspired by `fq` and `jq`. `jq` is a powerful and succinct query tool, sometimes I would find myself needing to use another bespoke tool for another format than json, whether its something dedicated with json query built in or a simple converter from one configuration format to json to pipe into jq. `qq` aims to be the only utility needed for most interaction with structured formats in the terminal. It can transcode configuration formats interchangeably between one-another with the power of `jq` and it has an `an interactive repl (with automcomplete)` to boot so you can have an interactive experience when building queries optionally. Many thanks to the authors of the libraries used in this project, especially `jq`, `gojq`, and `fq` for direct usage or inspiration for the project.


## Features
* support a wide range of configuration formats and transform them interchangeably between eachother.
* quick and comprehensive querying of configuration formats without needing a pipeline of dedicated tools.
* provide a fun to use interactive mode for building queries with autocomplete and realtime rendering preview.
* `qq` is broad, but focuses on performance of encodings (but mostly `gojq` is very fast), execution is often times faster than most any "jq but for `${x}` configuration format"-type tools. `qq` performs similarly to benchmarks of `jq` running on `JSON` itself in most covered formats.

consider the following file:
```shell
du -h large-file.json
25M     large-file.json
```

qq is not a slouch on performance despite covering a lot of codecs (disclaimer: this is only noticable with larger file sizes, and this hasn't been extensively compared.)

```shell
# gron large file bench

$ time gron large-file.json --no-sort | rg -v '[1-4]' | gron --ungron --no-sort > /dev/null 2>&1
gron large-file.json --no-sort  2.58s user 0.48s system 153% cpu 1.990 total
rg -v '[1-4]'  0.18s user 0.24s system 21% cpu 1.991 total
gron --ungron --no-sort > /dev/null 2>&1  7.68s user 1.15s system 197% cpu 4.475 total

$ time qq -o gron large-file.json | rg -v '[1-4]' | qq -i gron > /dev/null 2>&1
qq -o gron large-file.json  0.81s user 0.09s system 128% cpu 0.706 total
rg -v '[1-4]'  0.02s user 0.01s system 5% cpu 0.706 total
qq -i gron > /dev/null 2>&1  0.07s user 0.01s system 11% cpu 0.741 total
```

```shell
# yq large file bench

$ time yq large-file.json -M -o yaml > /dev/null 2>&1
yq large-file.json -M -o yaml > /dev/null 2>&1  4.02s user 0.31s system 208% cpu 2.081 total

$ time qq large-file.json -o yaml > /dev/null 2>&1
qq large-file.json -o yaml > /dev/null 2>&1  2.72s user 0.16s system 190% cpu 1.519 total
```

## Supported formats
Note: these unsupported formats are on a roadmap for inclusion.
| Format      | Input          | Output         |
|-------------|----------------|----------------|
| JSON        | ✅ Supported   | ✅ Supported   |
| YAML        | ✅ Supported   | ✅ Supported   |
| TOML        | ✅ Supported   | ✅ Supported   |
| XML         | ✅ Supported   | ✅ Supported   |
| INI         | ✅ Supported   | ✅ Supported   |
| HCL         | ✅ Supported   | ✅ Supported   |
| TF          | ✅ Supported   | ✅ Supported   |
| GRON        | ✅ Supported   | ✅ Supported   |
| CSV         | ✅ Supported   | ❌ Not Supported |
| Protobuf    | ❌ Not Supported | ❌ Not Supported |
| HTML        | ❌ Not Supported | ❌ Not Supported |


## Caveats
* `qq` is not a full `jq`/`*q` replacement and comes with idiosyncrasies from the underlying `gojq` library.
* the encoders and decoders are not perfect and may not be able to handle all edge cases.
* `qq` is under active development and more codecs are intended to be supported along with improvements to `interactive mode`.


## Contributions
All contributions are welcome to `qq`, especially for upkeep/optimization/addition of new encodings. For ideas on contributions [please refer to the todo docs](https://github.com/JFryy/qq/blob/main/docs/TODO.md) or make an issue/PR for a suggestion if there's something that's wanted or fixes.

## Thanks and Acknowledgements / Related Projects
* [gojq](https://github.com/itchyny/gojq): `gojq` is a pure Go implementation of jq. It is used to power the query engine of qq.
* [fq](https://github.com/wader/fq) : fq is a `jq` like tool for querying a wide array of binary formats.
* Many encoding modules 🍻
