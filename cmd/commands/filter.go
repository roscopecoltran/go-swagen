package commands

import (
	"errors"
	"fmt"
	"os"
	"path"

	flags "github.com/jessevdk/go-flags"
	"github.com/xreception/go-swagen/filter"
	"github.com/xreception/go-swagen/utils"
)

// Filter is a command that filter paths and definitions of a swagger file based on tags
type Filter struct {
	Input  flags.Filename `long:"input" short:"i" desciprtion:"input swagger file"`
	Output flags.Filename `long:"output" short:"o" description:"the file to write to"`
	Pretty bool           `long:"pretty" short:"p" description:"Prettify your output or not"`
	Tags   []string       `long:"tags" short:"t" description:"filter by tags"`
}

// Execute the command
func (c *Filter) Execute(args []string) error {
	if len(c.Input) == 0 {
		return errors.New("must have input file, plz use -i /path/to/swagger/file")
	}
	if len(c.Output) == 0 {
		// return errors.New("must define output directory, plz use -o")
		c.Output = "./build/swagger.json"
	}
	if _, err := os.Stat(string(c.Input)); os.IsNotExist(err) {
		return errors.New("input file does not exist")
	}
	dir := path.Dir(string(c.Output))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("# Creating output folder ...")
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
		fmt.Printf("# Folder %s is created.\n", dir)
	}

	fmt.Printf("# Starting filter process with tags %x ...\n", c.Tags)

	swagger, err := utils.LoadSpec(string(c.Input))
	if err != nil {
		return err
	}
	s := filter.Filter(swagger, c.Tags)
	err = utils.WriteToFile(s, c.Pretty, string(c.Output))
	if err != nil {
		return err
	}

	fmt.Println("# Filter Successfully!")

	return nil
}
