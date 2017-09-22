package filter

import (
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

	if utils.IsRef(s) {
		name := utils.GetRefName(s)
		ref := utils.GetRef(s, f.swagger)
		f.defs[name] = *ref
		f.schema(ref)
	}

	if utils.IsArray(s) {
		f.schema(s.Items.Schema)
	}

	if utils.IsObject(s) {
		for _, v := range s.Properties {
			f.schema(&v)
		}
	}
}
