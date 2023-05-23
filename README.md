Note that this is still under development. Feel free to check back soon!

# Introduction

`go2rascal` is a converter from Go ASTs to a Rascal-compatible AST format. 
This is designed to be used with the Go Analysis in Rascal (Go AiR) framework,
which can be found at [https://github.com/PLSE-Lab/go-analysis](https://github.com/PLSE-Lab/go-analysis).

# Running go2rascal

Although the intent is that this is run from Go AiR, you can also run this
directly. To do so, you can use the `go run` command, passing in the name
of the file to parse:

```
go run go2rascal.go --filePath /tmp/input.go
```

This command will parse the input script, given with the `--filePath`
command-line flag, using the parsing support that is included in 
standard Go libraries, and will then emit a textual version
of an AST for the file. This AST is in Rascal format, so it can be
pasted into a Rascal terminal that has already loaded the AST definition.

# Updates are Coming Soon!

We are at the very beginning of developing this, so it doesn't do much yet,
but please watch the repo and check back soon!