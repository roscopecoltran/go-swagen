package main

import (
	"github.com/go-openapi/spec"
	flags "github.com/jessevdk/go-flags"
	"github.com/xreception/go-swagen/factory"
	_ "github.com/xreception/go-swagen/generators/react_redux_typescript"
	"github.com/xreception/go-swagen/merger"
	"github.com/xreception/go-swagen/utils"
)

var opts struct {
	Inputs map[string]string `short:"i" long:"input" description:"Input scopes and files eg. -i scope:filepath"`
}

func main() {
	// make some fake args
	args := []string{
		"-i", "Account:./build/inputs/account.swagger.json",
		"-i", "Stock:./build/inputs/stock.swagger.json",
	}

	args, err := flags.ParseArgs(&opts, args)
	if err != nil {
		panic(err)
	}

	var scopes []string
	var swaggers []*spec.Swagger
	for scope, file := range opts.Inputs {
		scopes = append(scopes, scope)
		swagger, err := utils.LoadSpec(file)
		if err != nil {
			panic(err)
		}
		swaggers = append(swaggers, swagger)
	}

	output, err := merger.Merge(swaggers, scopes, nil, 1)
	if err != nil {
		panic(err)
	}
	err = utils.WriteToFile(output, true, "./build/gen/swagger.json")
	if err != nil {
		panic(err)
	}

	gen, err := factory.Create("react-redux-ts", map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	err = gen.Parse(output, "./build/gen")
	if err != nil {
		panic(err)
	}
}
