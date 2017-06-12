package reactReduxTypescript

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"runtime"

	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/xreception/go-swagen/factory"
	"github.com/xreception/go-swagen/utils"
)

const generatorName = "react-redux-ts"
const entityID = "uri"

func init() {
	factory.Register(generatorName, &theFactory{})
}

// theFactory implements factory.IFactory
type theFactory struct{}

func (f *theFactory) Create(parameters map[string]interface{}) (factory.IGenerator, error) {
	return &generator{
		actions:   make(map[string][]*Action),
		hasSchema: make(map[string]bool),
	}, nil
}

// generator implements factory.IGenerator
type generator struct {
	factory.IGenerator
	swagger *spec.Swagger
	schemas []*Schema
	actions map[string][]*Action

	hasSchema map[string]bool
}

// Schema the normalizr Schema structure
type Schema struct {
	Name  string
	Class string
	Deps  map[string]bool
}

// Action the readux Action
type Action struct {
	Name       string
	Type       string
	Method     string
	Endpoint   string
	SchemaName string
	Parameters []spec.Parameter
}

// Parse implements IGenerator's Parse method.
func (gen *generator) Parse(swagger *spec.Swagger, out string) error {
	gen.swagger = swagger

	paths := gen.swagger.Paths
	if paths == nil || paths.Paths == nil || len(paths.Paths) == 0 {
		return errors.New("this swagger has no paths")
	}

	for endpoint, item := range paths.Paths {
		gen.parseOperation(endpoint, "GET", item.Get)
		gen.parseOperation(endpoint, "PUT", item.Put)
		gen.parseOperation(endpoint, "POST", item.Post)
	}

	return gen.writeTo(out)
}

// ParseFile implements IGenerator's ParseFile method
func (gen *generator) ParseFile(in string, out string) error {
	doc, err := loads.Spec(in)
	if err != nil {
		return err
	}

	return gen.Parse(doc.Spec(), out)
}

func (gen *generator) writeTo(folder string) error {
	m := map[string]interface{}{"action": gen.actions, "api": gen.actions, "constant": gen.actions, "schema": gen.schemas}
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return errors.New("could not open default swagger.json file")
	}
	dir := path.Dir(filename)

	for k, v := range m {
		file := fmt.Sprintf("%s/%s.ts", folder, k)
		tmpl := path.Join(dir, fmt.Sprintf("./templates/%s.tmpl", k))
		err := writeFile(file, tmpl, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFile(filePath string, tmplPath string, data interface{}) error {
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		return err
	}

	tmpl := template.Must(template.ParseFiles(tmplPath)).Funcs(template.FuncMap{
		"CamelCase": utils.CamelCase,
	})
	return tmpl.Execute(file, data)
}

// parseOperation parse the operation of swagger.
func (gen *generator) parseOperation(endpoint string, method string, op *spec.Operation) {
	if op == nil {
		return
	}

	if op.Responses == nil || op.Responses.StatusCodeResponses == nil {
		return
	}

	// TODO (junhua): only pass response with 200 for now
	if resp, ok := op.Responses.StatusCodeResponses[200]; ok {
		// TODO(junhua): suppose every response is a ref schema
		schemaName := getSchemaName(resp.Schema)
		gen.parseSchema(resp.Schema, schemaName)
		a := &Action{
			Name:       utils.CamelCase(op.ID),
			Type:       utils.SnakeCase(op.Tags[0] + op.ID),
			Method:     method,
			Endpoint:   replaceWithDollar(endpoint),
			SchemaName: schemaName,
			Parameters: op.Parameters,
		}

		for _, s := range op.Tags {
			s = utils.CamelCase(s)
			as, ok := gen.actions[s]
			if !ok {
				as = make([]*Action, 0)
			}

			as = append(as, a)
			gen.actions[s] = as
		}
	}
}

// parseSchema parse the schema.
func (gen *generator) parseSchema(s *spec.Schema, name string) {
	if s == nil || gen.hasSchema[name] || len(name) == 0 {
		return
	}

	if s.Ref.HasFragmentOnly {
		data, _, err := s.Ref.GetPointer().Get(gen.swagger)
		if err != nil {
			fmt.Println(err) // TODO(junhua) 这里的错误可以安全吞掉
		}
		if data == nil {
			return
		}

		nextSchema := data.(spec.Schema)
		gen.parseSchema(&nextSchema, name)
		return
	}

	schema := &Schema{
		Name: name,
		Deps: make(map[string]bool),
	}
	schema.Class = "Object"
	for k, v := range s.Properties {
		if k == entityID {
			schema.Class = "Entity"
		}
		if v.Ref.HasFragmentOnly {
			refName := getSchemaName(&v)
			gen.parseSchema(&v, refName)
			schema.Deps[refName] = true
		}
		if isArrayOfSchema(&v) {
			refName := getSchemaName(v.Items.Schema)
			gen.parseSchema(v.Items.Schema, refName)
			schema.Deps[refName] = true
		}
	}
	gen.schemas = append(gen.schemas, schema)
	gen.hasSchema[name] = true
}

func getSchemaName(s *spec.Schema) string {
	pr := s.Ref.GetPointer()
	tokens := pr.DecodedTokens()
	return utils.CamelCase(tokens[len(tokens)-1] + "Schema")
}

func isArrayOfSchema(s *spec.Schema) bool {
	return s.Items != nil && s.Items.Schema != nil && s.Items.Schema.Ref.HasFragmentOnly
}

func replaceWithDollar(s string) string {
	return strings.Replace(s, "{", "${", -1)
}
