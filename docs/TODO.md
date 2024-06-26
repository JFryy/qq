## TODO

* Support for GRON (!!) - maybe include regex engine for direct conversion without intermediate path format? (see how much bloat it would add at least). This could be done with an -E (or -G) flag and -Ev to filter out pattern matches.
* push to homebrew and AUR and dockerhub
* Support for HTML
* Support for excel family
* TUI View fixes on large files
* TUI Autocompletion improvements (back/forward/based on partial content of path rather than dorectly iterating through splatted gron like paths)
* csv codec improvements: list of maps by default, more agressive heurestics for parsing.
* Support slurp and many other flags of jq that are useful.
* Support for protobuff
* more complex tests (but still keep the cli tests) with post-conversion/query type assertions
