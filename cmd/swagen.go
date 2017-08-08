package main

import (
	"log"
	"os"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	"github.com/jessevdk/go-flags"
	"github.com/xreception/go-swagen/cmd/commands"
	_ "github.com/xreception/go-swagen/generators/react_redux_typescript"
	_ "github.com/xreception/go-swagen/generators/typescript"
)

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
}

var opts struct {
	// Version bool `long:"version" short:"v" description:"print the version of the command"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.ShortDescription = "helps you keep your API well described"
	parser.LongDescription = `
Swagen tries to support you as best as possible when building APIs.
Merge multiple swagger files into one.
Generate js or ts client from given swagger file.
`
	_, err := parser.AddCommand("merge", "merge files", "merge multiple swagger files into one", &commands.Merge{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = parser.AddCommand("generate", "generate client sdk", "generate client sdk for given language", &commands.Generate{})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
