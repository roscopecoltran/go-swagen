package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

// LoadSpec loads the swagger.json file
func LoadSpec(input string) (*spec.Swagger, error) {
	if fi, err := os.Stat(input); err == nil {
		if fi.IsDir() {
			return nil, fmt.Errorf("expected %q to be a file not a directory", input)
		}
		sp, err := loads.Spec(input)
		if err != nil {
			return nil, err
		}
		return sp.Spec(), nil
	}
	return nil, nil
}

// WriteToFile dump inmemory swagger to file
func WriteToFile(swspec *spec.Swagger, pretty bool, output string) error {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(swspec, "", "  ")
	} else {
		b, err = json.Marshal(swspec)
	}
	if err != nil {
		return err
	}
	if output == "" {
		fmt.Println(string(b))
		return nil
	}
	return ioutil.WriteFile(output, b, 0644)
}

// LoadSpecsWithScopes load swagger specs from []string.
// Each item follow format scope@filepath.
func LoadSpecsWithScopes(inputs []string) ([]*spec.Swagger, []string, error) {
	var scopes []string
	var swaggers []*spec.Swagger
	for _, input := range inputs {
		scope := ""
		file := input
		ss := strings.Split(input, "@")
		if len(ss) > 2 {
			return nil, nil, errors.New("at most one @ character in inputs")
		}
		if len(ss) == 2 {
			scope = ss[0]
			file = ss[1]
		}

		scopes = append(scopes, scope)
		swagger, err := LoadSpec(file)
		if err != nil {
			return nil, nil, err
		}
		swaggers = append(swaggers, swagger)
	}

	return swaggers, scopes, nil
}
