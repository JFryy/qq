# qq

[![Go](https://github.com/JFryy/qq/actions/workflows/go.yml/badge.svg)](https://github.com/JFryy/qq/actions/workflows/go.yml)
[![Docker Build](https://github.com/JFryy/qq/actions/workflows/docker-image.yml/badge.svg)](https://github.com/JFryy/qq/actions/workflows/docker-image.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/JFryy/qq)](https://golang.org/)
[![License](https://img.shields.io/github/license/JFryy/qq)](https://github.com/JFryy/qq/blob/main/LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/JFryy/qq)](https://github.com/JFryy/qq/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/jfryy/qq)](https://hub.docker.com/r/jfryy/qq)

`qq` is an interoperable configuration format transcoder with `jq` query syntax powered by `gojq`. `qq` is multi modal, and can be used as a replacement for `jq` or be interacted with via 
a REPL with autocomplete and realtime rendering preview for building queries.

`qq` is designed to support input output operations on a large variety of structured data codecs with the power of jq, please refer to the below for supported formats/extensions.

## Usage

Here's some example usage, this emphasizes the interactive mode for demonstration, but `qq` is designed for usage in shell scripts.
![Demo GIF](docs/demo.gif)

```sh
# JSON is default in and output.
cat file.${ext} | qq -i ${ext}

# Extension is parsed, no need for input flag
qq '.' file.xml

# random example: query xml, grep with gron using qq io and output as json
qq file.xml -o gron | grep -vE "sweet.potatoes" | qq -i gron

# get some content from a site with html input
curl motherfuckingwebsite.com | bin/qq -i html '.html.body.ul.li[0]'

# interactive query builder mode on target file
qq . file.json --interactive

# streaming mode - works with JSON, YAML, CSV and more
qq --stream 'select(length == 2)' large.json

# slurp mode - read multiple inputs into an array
echo -e '{"id":1}\n{"id":2}' | qq -s 'map(.id)'

# exit-status - use in conditionals
echo '{"active":true}' | qq -e '.active' && echo "is active"
```

## Git

You can also use it for cleaner diffing of configuration files by adding to your `git/config` file a snippet such as

```
  [diff "csv"]
  textconv = "f(){ in=\"$1\"; \
      if command -v qq   >/dev/null 2>&1; then qq --monochrome-output --output gron --input csv  \"$in\" 2>/dev/null | sort && exit 0; fi; \
      cat \"$in\"; \
    }; f"
  [diff "env"]
    textconv = "f(){ qq --monochrome-output --output gron --input env  \"$1\" 2>/dev/null | sort || cat \"$1\"; }; f"
  [diff "html"]
    textconv = "f(){ qq --monochrome-output --output gron --input html \"$1\" 2>/dev/null | sort || cat \"$1\"; }; f"
  [diff "ini"]
    textconv = "f(){ qq --monochrome-output --output gron --input ini  \"$1\" 2>/dev/null | sort || cat \"$1\"; }; f"
  [diff "toml"]
    textconv = "f(){ qq --monochrome-output --output gron --input toml \"$1\" 2>/dev/null | sort || cat \"$1\"; }; f"
```

and to `git/attributes` correspondingly

```
*.csv diff=csv
*.env diff=env
*.html diff=html
*.ini diff=ini
*.toml diff=toml
```


## Installation

From brew:

```shell
brew install jfryy/tap/qq 
```

From [AUR](https://aur.archlinux.org/packages/qq-git) (ArchLinux):

```shell
yay qq-git
```

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

`qq` is inspired by `fq` and `jq`. `jq` is a powerful and succinct query tool, sometimes I would find myself needing to use another bespoke tool for another format than json, whether its something dedicated with json query built in or a simple converter from one configuration format to json to pipe into jq. `qq` aims to be a handly utility on the terminal or in shell scripts that can be used for most interaction with structured formats in the terminal. It can transcode configuration formats interchangeably between one-another with the power of `jq` and it has an `an interactive repl (with automcomplete)` to boot so you can have an interactive experience when building queries optionally. Many thanks to the authors of the libraries used in this project, especially `jq`, `gojq`, `gron` and `fq` for direct usage and/or inspiration for the project.

## Features

* Support a wide range of configuration formats and transform them interchangeably between each other.
* Quick and comprehensive querying of configuration formats without needing a pipeline of dedicated tools.
* Provide an interactive mode for building queries with autocomplete and realtime rendering preview.
* Streaming mode (`--stream`) (identical to jq's `--stream`), plus extended support for JSONL, YAML, CSV, TSV, and line-delimited formats - all emit path-value pairs for memory-efficient processing of large files.
* `qq` is broad, but performant encodings are still a priority, execution is quite fast despite covering a broad range of codecs. `qq` performs comparitively with dedicated tools for a given format.


## Supported Extensions/Formats

```
.json, .jsonl, .ndjson, .jsonlines, .jsonc, .yaml, .yml, .toml, .xml, .ini, .hcl, .tf,
.gron, .csv, .tsv, .properties, .html, .parquet, .msgpack, .mpk, .base64, .b64,
.proto (input only), .txt (input only), .env (input only)
```

## Contributions

All contributions are welcome to `qq`, especially for upkeep/optimization/addition of new encodings.

## Thanks and Acknowledgements / Related Projects

This tool would not be possible without the following projects, this project is arguably more of a composition of these projects than a truly original work, with glue code, some dedicated encoders/decoders, and the interactive mode being original work.
Nevertheless, I hope this project can be useful to others, and I hope to contribute back to the community with this project.

* [gojq](https://github.com/itchyny/gojq): `gojq` is a pure Go implementation of jq. It is used to power the query engine of qq.
* [fq](https://github.com/wader/fq) : fq is a `jq` like tool for querying a wide array of binary formats.
* [jq](https://github.com/jqlang/jq): `jq` is a lightweight and flexible command-line JSON processor.
* [gron](https://github.com/tomnomnom/gron): gron transforms JSON into discrete assignments that are easy to grep.
* [yq](https://github.com/mikefarah/yq): yq is a lightweight and flexible command-line YAML (and much more) processor.
* [goccy](https://github.com/goccy/go-json): goccy has quite a few encoders and decoders for various formats, and is used in the project for some encodings.
* [go-toml](https://github.com/BurntSushi/toml): go-toml is a TOML parser for Golang with reflection.
