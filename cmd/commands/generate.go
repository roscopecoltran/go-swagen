package commands

import (
	"errors"

	"os"

	"fmt"

	flags "github.com/jessevdk/go-flags"
	"github.com/xreception/go-swagen/factory"
	"github.com/xreception/go-swagen/utils"
)

// Generate is a command that merge multiple files into one swagger document
type Generate struct {
	Lang   string         `long:"lang" short:"l" description:"target language of client sdk"`
	Input  flags.Filename `long:"input" short:"i" desciprtion:"input swagger files, you could use scope@filename if want to put a scope for the swagger"`
	Output string         `long:"output" short:"o" description:"the path to write to"`
}

// Execute expands the spec
func (c *Generate) Execute(args []string) error {
	if len(args) != 0 {
		c.Input = flags.Filename(args[0])
	}
	// validation flags
	if len(c.Input) == 0 {
		return errors.New("must have input, plz use -i")
	}
	if len(c.Output) == 0 {
		// return errors.New("must define output directory, plz use -o")
		c.Output = "./build/gen"
	}
	if len(c.Lang) == 0 {
		c.Lang = "react-redux-ts" // now we only support react-redux-ts lang
		// return errors.New("Plz define the target language to generate specific client sdk, use -l")
	}
	if _, err := os.Stat(string(c.Input)); os.IsNotExist(err) {
		return errors.New("input file does not exist")
	}
	if _, err := os.Stat(c.Output); os.IsNotExist(err) {
		fmt.Println("# Creating output folder ...")
		err := os.MkdirAll(c.Output, os.ModePerm)
		if err != nil {
			return err
		}
		fmt.Printf("# Folder %s is created.\n", c.Output)
	}

	fmt.Printf("# Generating %s code ...\n", c.Lang)
	swagger, err := utils.LoadSpec(string(c.Input))
	if err != nil {
		return err
	}
	gen, err := factory.Create(c.Lang, map[string]interface{}{})
	if err != nil {
		return err
	}
	err = gen.Parse(swagger, c.Output)
	if err != nil {
		panic(err)
	}

	fmt.Println("# Generated Successfully!")

	return nil
}
