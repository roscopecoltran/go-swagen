package typescript

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"text/template"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/xreception/go-swagen/factory"
	"github.com/xreception/go-swagen/generators"
	"github.com/xreception/go-swagen/utils"
)

const generatorName = "typescript"
const entityID = "uri"

var templates *generators.Repository
var assets = map[string][]byte{
	"schema.tmpl":  MustAsset("templates/schema.tmpl"),
	"service.tmpl": MustAsset("templates/service.tmpl"),
	"request.tmpl": MustAsset("templates/request.tmpl"),
}

// FuncMap is a map with default functions for use n the templates.
// These are available in every template
var FuncMap template.FuncMap = map[string]interface{}{
	"CamelCase":      utils.CamelCase,
	"InterfaceCase":  utils.InterfaceCase,
	"PluralCase":     utils.PluralCase,
	"UpperSnakeCase": utils.UpperSnakeCase,
	"schemaType":     schemaType,
	"schemaRef":      schemaRef,
	"parameterType":  parameterType,
	"responseType":   responseType,
}

func init() {
	factory.Register(generatorName, &theFactory{})
	templates = generators.NewRepository(FuncMap)
	templates.LoadAssets(assets)
}

// theFactory implements factory.IFactory
type theFactory struct{}

func (f *theFactory) Create(parameters map[string]interface{}) (factory.IGenerator, error) {
	return &generator{
		Schemas:  make([]*spec.Schema, 0),
		Services: make(map[string][]*spec.Operation),
	}, nil
}

// generator implements factory.IGenerator
type generator struct {
	factory.IGenerator
	swagger *spec.Swagger

	Schemas  []*spec.Schema
	Services map[string][]*spec.Operation
}

// Parse implements IGenerator's Parse method.
func (gen *generator) Parse(swagger *spec.Swagger, out string) error {
	gen.swagger = swagger

	paths := gen.swagger.Paths
	if paths == nil || paths.Paths == nil || len(paths.Paths) == 0 {
		return errors.New("this swagger has no path")
	}

	for _, name := range utils.SortedStringKeys(swagger.Definitions) {
		schema := swagger.Definitions[name]
		schema.ID = name
		gen.Schemas = append(gen.Schemas, &schema)
	}

	for _, endpoint := range utils.SortedStringKeys(paths.Paths) {
		item := paths.Paths[endpoint]
		gen.parseOperation(endpoint, "GET", item.Get)
		gen.parseOperation(endpoint, "PUT", item.Put)
		gen.parseOperation(endpoint, "POST", item.Post)
		gen.parseOperation(endpoint, "DELETE", item.Delete)
	}

	return gen.write(out)
}

// ParseFile implements IGenerator's ParseFile method
func (gen *generator) ParseFile(in string, out string) error {
	doc, err := loads.Spec(in)
	if err != nil {
		return err
	}

	return gen.Parse(doc.Spec(), out)
}

func (gen *generator) write(folder string) error {
	err := gen.writeAPI(folder)
	if err != nil {
		return err
	}

	err = gen.writeSchema(folder)
	if err != nil {
		return err
	}

	return gen.writeRequest(folder)
}

func (gen *generator) writeAPI(folder string) error {
	for service, operations := range gen.Services {
		filePath := fmt.Sprintf("%s/%s.ts", folder, service)
		file, err := os.Create(filePath)
		defer file.Close()
		if err != nil {
			return err
		}

		err = templates.ExecuteTemplate(file, "service", struct {
			Service    string
			Operations []*spec.Operation
		}{
			service,
			operations,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (gen *generator) writeSchema(folder string) error {
	filePath := fmt.Sprintf("%s/%s.ts", folder, "schema")
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		return err
	}

	return templates.ExecuteTemplate(file, "schema", gen)
}

func (gen *generator) writeRequest(folder string) error {
	filePath := fmt.Sprintf("%s/%s.ts", folder, "request")
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		return err
	}

	return templates.ExecuteTemplate(file, "request", gen)
}

// parseOperation parse the operation of swagger.
func (gen *generator) parseOperation(endpoint string, method string, op *spec.Operation) {
	if op == nil {
		return
	}

	op.AddExtension("method", method)
	op.AddExtension("endpoint", replaceWithDollar(endpoint))

	for i := range op.Parameters {
		gen.parseParam(&op.Parameters[i])
	}

	for _, tag := range op.Tags {
		operations, ok := gen.Services[tag]
		if !ok {
			operations = make([]*spec.Operation, 0)
		}

		operations = append(operations, op)
		gen.Services[tag] = operations
	}
}

// parseParam parse the parameters of operation
func (gen *generator) parseParam(param *spec.Parameter) {
	if param.Schema != nil {
		param.Type = utils.InterfaceCase(param.Schema.ID)
	} else if param.Type == "integer" {
		param.Type = "number"
	}
}

func replaceWithDollar(s string) string {
	re := regexp.MustCompile(`{[a-z0-9A-Z_-]+}`)
	return re.ReplaceAllStringFunc(s, func(matched string) string {
		n := len(matched)
		w := matched[1 : n-1]
		return "${" + utils.CamelCase(w) + "}"
	})
}
