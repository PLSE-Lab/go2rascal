# Introduction

`go2rascal` is a converter from Go ASTs to a Rascal-compatible AST format. 
This is designed to be used with the Go Analysis in Rascal (Go AiR) framework,
which can be found at [https://github.com/PLSE-Lab/go-analysis](https://github.com/PLSE-Lab/go-analysis).

# Running go2rascal

Although the intent is that this is run from Go AiR, you can also run this
directly. To do so, you can use the `go run` command, passing in the name
of the file or project to parse. For an individual file, you would run:

```
go run go2rascal.go --filePath /tmp/input.go
```

This command will parse the input script, given with the `--filePath`
command-line flag, using the parsing support that is included in
standard Go libraries, and will then emit a textual version
of an AST for the file. This AST can then be saved into a file to load later, pasted into a Rascal console, or pasted into a Rascal source file.

For an entire project, you would run:

```
go run go2rascal.go --projectPath /path/to/project/ --scratchDir /path/tp/results
```

Since there are multiple files, a Rascal map from the original source file to a scratch file with the AST is generated. The scratch directory then contains the AST files, which can be loaded into Rascal. Note that the scratch directory is deleted and recreated every time this is run. Note that this uses the [Go packages library](https://pkg.go.dev/golang.org/x/tools/go/packages) to handle multiple modules inside a single project, and will only process the source files that would be processed by a build of that module (for instance, additional modules given in subfolders would not be included).

You can enable and disable the addition of source locations on the generated
ASTs using the `--addLocs` flag. By default, locations are added, which is
equivalent to `--addLocs=true`. Passing the flag `-addLocs=false` will disable
the addition of source locations. An example of this would be:

```
go run go2rascal.go --filePath /tmp/input.go --addLocs=false
```

Another flag, `--addTypes`, is used to include type information in the generated ASTs. Note that this _can_ work when parsing a single file, but can easily fail to work since files often include external module dependencies. If type information is needed it is better to process the entire project, since the package library will process the files in the same way as the build system and will properly resolve types from external modules.

Specifically for processing projects, the `--includeTests` flag includes test files. This can be used if the ASTs for test files are needed. If this is not included, test files are not processed since they would not be processed in a build.

An example command, putting all of these together, is:
```
go run go2rascal.go --addLocs=true --addTypes=true --includeTests=true --projectPath /somedir/systems/hashicorp-raft --scratchDir /somedir/scratch
```

Finally, to speed up the AST generation process, you can build the `go2rascal`
binary using the command `go build go2rascal.go`. This will create an
executable named `go2rascal` in the current directory. If the `runConverterBinary`
flag in the Go AiR configuration is set to `true`, this binary will
be used instead of running `go2rascal.go` using `go run`, which compiles
the file each time. Running the file from the binary uses the same arguments
as running it using `go run`, for instance:

```
./go2rascal --filePath /tmp/input.go --addLocs=false
```

While this makes very little difference when processing a single file or system, running the binary can result in a significant performance boost when executed repeatedly (e.g., when processing a large amount of Go code across multiple systems).