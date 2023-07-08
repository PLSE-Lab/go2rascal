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

You can enable and disable the addition of source locations on the generated
ASTs using the `--addLocs` flag. By default, locations are added, which is
equivalent to `--addLocs=True`. Passing the flag `-addLocs=False` will disable
the addition of source locations. An example of this would be:

```
go run go2rascal.go --filePath /tmp/input.go --addLocs=False
```

Finally, to speed up the AST generation process, you can build the `go2rascal`
binary using the command `go build go2rascal.go`. This will create an
executable named `go2rascal` in the current directory. If the `runConverterBinary`
flag in the Go AiR configuration is set to `true`, this binary will
be used instead of running `go2rascal.go` using `go run`, which compiles
the file each time.
