package reactReduxTypescript

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"runtime"
	"text/template"

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
		Actions:   make(map[string][]*Action),
		hasSchema: make(map[string]*Schema),
	}, nil
}

// generator implements factory.IGenerator
type generator struct {
	factory.IGenerator
	swagger   *spec.Swagger
	hasSchema map[string]*Schema

	Actions map[string][]*Action
	Schemas []*Schema
}

// Schema the normalizr Schema structure
type Schema struct {
	Name         string
	Class        string
	Deps         map[string]string
	Normalizable bool
	Props        map[string]string
	Enum         []interface{}
}

// Action the readux Action
type Action struct {
	Name       string
	Type       string
	Method     string
	Endpoint   string
	RespSchema *Schema
	Parameters []spec.Parameter
}

// Parse implements IGenerator's Parse method.
func (gen *generator) Parse(swagger *spec.Swagger, out string) error {
	gen.swagger = swagger

	paths := gen.swagger.Paths
	if paths == nil || paths.Paths == nil || len(paths.Paths) == 0 {
		return errors.New("this swagger has no path")
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
	// locate template folder
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return errors.New("could not get runtime file name")
	}
	dir := path.Dir(filename)
	tmplDir := path.Join(dir, fmt.Sprintf("./templates"))

	// prepare templates
	funcMap := template.FuncMap{
		"CamelCase":     utils.CamelCase,
		"InterfaceCase": utils.InterfaceCase,
		"PluralCase":    utils.PluralCase,
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob(tmplDir + "/*.tmpl"))

	// i/o writer
	m := map[string]interface{}{"action": gen.Actions, "api": gen, "constant": gen.Actions, "schema": gen.Schemas}
	for k, v := range m {
		filePath := fmt.Sprintf("%s/%s.ts", folder, k)
		file, err := os.Create(filePath)
		defer file.Close()
		if err != nil {
			return err
		}

		err = tmpl.ExecuteTemplate(file, k+".tmpl", v)
		if err != nil {
			return err
		}
	}
	return nil
}

// parseOperation parse the operation of swagger.
func (gen *generator) parseOperation(endpoint string, method string, op *spec.Operation) {
	if op == nil {
		return
	}

	for i := range op.Parameters {
		gen.parseParam(&op.Parameters[i])
	}

	if op.Responses == nil || op.Responses.StatusCodeResponses == nil {
		return
	}

	// TODO (junhua): only pass response with 200 for now
	if resp, ok := op.Responses.StatusCodeResponses[200]; ok {
		// TODO(junhua): suppose every response is a ref schema
		schemaName := getSchemaName(resp.Schema)
		schema := gen.parseSchema(resp.Schema, schemaName)
		a := &Action{
			Name:       utils.CamelCase(op.ID),
			Type:       utils.UpperSnakeCase(op.Tags[0] + op.ID),
			Method:     method,
			Endpoint:   replaceWithDollar(endpoint),
			Parameters: op.Parameters,
		}

		if schema != nil {
			a.RespSchema = schema
		}

		for _, s := range op.Tags {
			s = utils.CamelCase(s)
			as, ok := gen.Actions[s]
			if !ok {
				as = make([]*Action, 0)
			}

			as = append(as, a)
			gen.Actions[s] = as
		}
	}
}

// parseParam parse the parameters of operation
func (gen *generator) parseParam(param *spec.Parameter) {
	if param.Schema != nil {
		name := getSchemaName(param.Schema)
		gen.parseSchema(param.Schema, name)
		param.Type = utils.InterfaceCase(name)
	} else if param.Type == "integer" {
		param.Type = "number"
	}
}

// parseSchema parse the schema.
func (gen *generator) parseSchema(s *spec.Schema, name string) *Schema {
	if s == nil || len(name) == 0 {
		return nil
	}

	if existed, ok := gen.hasSchema[name]; ok {
		return existed
	}

	if s.Ref.HasFragmentOnly {
		data, _, err := s.Ref.GetPointer().Get(gen.swagger)
		if err != nil {
			fmt.Println(err) // TODO(junhua) 这里的错误可以安全吞掉
		}
		if data == nil {
			return nil
		}

		nextSchema := data.(spec.Schema)
		return gen.parseSchema(&nextSchema, name)
	}

	schema := &Schema{
		Name:         name,
		Deps:         make(map[string]string),
		Props:        make(map[string]string),
		Class:        "Object",
		Normalizable: false,
		Enum:         s.Enum,
	}
	for k, v := range s.Properties {
		schema.Props[k] = getSchemaType(&v)

		if k == entityID {
			schema.Class = "Entity"
			schema.Normalizable = true
		} else if v.Ref.HasFragmentOnly {
			refName := getSchemaName(&v)
			schema.Props[k] = utils.InterfaceCase(refName)
			next := gen.parseSchema(&v, refName)
			if next.Normalizable {
				schema.Deps[k] = utils.CamelCase(refName)
				schema.Normalizable = true
			}
		} else if isArrayOfSchema(&v) {
			refName := getSchemaName(v.Items.Schema)
			schema.Props[k] = utils.InterfaceCase(refName) + "[]"
			next := gen.parseSchema(v.Items.Schema, refName)
			if next.Normalizable {
				schema.Deps[k] = "[" + utils.CamelCase(refName) + "]"
				schema.Normalizable = true
			}
		}
	}

	gen.Schemas = append(gen.Schemas, schema)
	gen.hasSchema[name] = schema
	return schema
}

func getSchemaName(s *spec.Schema) string {
	pr := s.Ref.GetPointer()
	tokens := pr.DecodedTokens()
	return tokens[len(tokens)-1]
}

func getSchemaType(schema *spec.Schema) string {
	if len(schema.Type) == 0 {
		return ""
	}

	t := schema.Type[0]
	a := ""
	if t == "array" && len(schema.Items.Schema.Type) > 0 {
		t = schema.Items.Schema.Type[0]
		a = "[]"
	}
	if t == "integer" {
		t = "number"
	}
	return t + a
}

func isArrayOfSchema(s *spec.Schema) bool {
	return s.Items != nil && s.Items.Schema != nil && s.Items.Schema.Ref.HasFragmentOnly
}

func replaceWithDollar(s string) string {
	re := regexp.MustCompile(`{[a-z0-9A-Z_-]+}`)
	return re.ReplaceAllStringFunc(s, func(matched string) string {
		n := len(matched)
		w := matched[1 : n-1]
		return "${" + utils.CamelCase(w) + "}"
	})
}
