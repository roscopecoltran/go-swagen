package filter

import (
	"log"
	"reflect"

	"github.com/go-openapi/spec"
	"github.com/xreception/go-swagen/utils"
)

// Filter the swagger file based on given tags
func Filter(swagger *spec.Swagger, tags []string) *spec.Swagger {
	f := &filter{
		swagger: swagger,
		tags:    tags,
		paths:   make(map[string]spec.PathItem),
		defs:    make(map[string]spec.Schema),
	}

	return f.Run()
}

type filter struct {
	swagger *spec.Swagger
	tags    []string

	paths map[string]spec.PathItem
	defs  spec.Definitions
}

// Run start
func (f *filter) Run() *spec.Swagger {
	for endpoint, path := range f.swagger.Paths.Paths {
		f.path(endpoint, path)
	}

	s := &spec.Swagger{
		SwaggerProps: f.swagger.SwaggerProps,
	}
	s.Paths.Paths = f.paths
	s.Definitions = f.defs

	return s
}

func (f *filter) path(endpoint string, path spec.PathItem) {
	p := reflect.ValueOf(path)
	if !p.IsValid() {
		return
	}

	var toBeAdd bool
	for _, method := range []string{"Get", "Put", "Post", "Delete", "Options", "Head", "Patch"} {
		v := reflect.Indirect(p).FieldByName(method)
		if !v.IsValid() {
			continue
		}
		op := v.Interface().(*spec.Operation)
		if op == nil {
			continue
		}

		ct := utils.Intersection(op.Tags, f.tags)
		if len(ct) == 0 {
			indir := v.Elem()
			indir.Set(reflect.Zero(indir.Type()))
		} else {
			op.Tags = ct
			toBeAdd = true
			f.operation(op)
		}
	}

	if toBeAdd {
		f.paths[endpoint] = path
	}
}

func (f *filter) operation(op *spec.Operation) {
	if op == nil {
		return
	}

	for _, p := range op.Parameters {
		f.schema(p.Schema)
	}

	if op.Responses == nil {
		return
	}

	f.response(op.Responses.Default)
	for _, v := range op.Responses.StatusCodeResponses {
		f.response(&v)
	}
}

func (f *filter) response(res *spec.Response) {
	if res == nil {
		return
	}

	f.schema(res.Schema)
}

func (f *filter) schema(s *spec.Schema) {
	if s == nil {
		return
	}

	if isRef(s) {
		name := getRefName(s)
		f.defs[name] = *getRef(s, f.swagger)
		f.schema(getRef(s, f.swagger))
	}

	if isArray(s) {
		f.schema(s.Items.Schema)
	}

	if isObject(s) {
		for _, v := range s.Properties {
			f.schema(&v)
		}
	}
}

func getRefName(s *spec.Schema) string {
	pr := s.Ref.GetPointer()
	tokens := pr.DecodedTokens()
	return tokens[len(tokens)-1]
}

func getRef(s *spec.Schema, document interface{}) *spec.Schema {
	data, _, err := s.Ref.GetPointer().Get(document)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	refSchema := data.(spec.Schema)
	return &refSchema
}

func isArray(schema *spec.Schema) bool {
	return utils.Contains(schema.Type, "array")
}

func isObject(schema *spec.Schema) bool {
	return utils.Contains(schema.Type, "object")
}

func isRef(s *spec.Schema) bool {
	return s.Ref.HasFragmentOnly
}
