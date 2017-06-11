package commands

import (
	"errors"
	"os"
	"path"

	flags "github.com/jessevdk/go-flags"
	"github.com/xreception/go-swagen/merger"
	"github.com/xreception/go-swagen/utils"
)

// Merge is a command that merge multiple files into one swagger document
type Merge struct {
	CompressLevel int            `long:"compress" short:"c" description:"compress level"`
	Inputs        []string       `long:"input" short:"i" desciprtion:"input swagger files, you could use scope@filename if want to put a scope for the swagger"`
	Output        flags.Filename `long:"output" short:"o" description:"the file to write to"`
	Pretty        bool           `long:"pretty" short:"p" description:"Prettify your output or not"`
}

// Execute expands the spec
func (c *Merge) Execute(args []string) error {
	for _, fileFromArg := range args {
		c.Inputs = append(c.Inputs, fileFromArg)
	}

	// validation flags
	if c.CompressLevel < 0 {
		return errors.New("compress level should not lower than 0")
	}
	if len(c.Inputs) == 0 {
		return errors.New("must have inputs, plz use -i")
	}
	if len(c.Output) == 0 {
		// return errors.New("must define output directory, plz use -o")
		c.Output = "./build/swagger.json"
	}
	dir := path.Dir(string(c.Output))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	swaggers, scopes, err := utils.LoadSpecsWithScopes(c.Inputs)
	if err != nil {
		return err
	}
	output, err := merger.Merge(swaggers, scopes, nil, c.CompressLevel)
	if err != nil {
		return err
	}
	err = utils.WriteToFile(output, c.Pretty, string(c.Output))
	if err != nil {
		return err
	}

	return nil
}
