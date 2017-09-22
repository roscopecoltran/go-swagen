package merger

import (
	"bytes"
	"crypto/md5"
	"errors"
	"log"

	"github.com/go-openapi/spec"
	"github.com/xreception/go-swagen/utils"
)

// IMarshaler interface
type IMarshaler interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

type merger struct {
	primary    *spec.Swagger
	defs       spec.Definitions
	paths      map[string]spec.PathItem
	revertDefs map[string]string
	replaceMap map[string]string
}

var prefix = "#/definitions/"
var suffix = "\""

// Merge multiple swaggers to one swagger
func Merge(swaggers []*spec.Swagger, scopes []string, primary *spec.Swagger, compressLevel int) (*spec.Swagger, error) {
	var err error
	if len(swaggers) == 0 {
		return primary, nil
	}
	if len(swaggers) != len(scopes) {
		return nil, errors.New("Lenght of scopes and swaggers should be equal")
	}
	if primary == nil {
		primary = defaultSwagger()
	}
	if primary.Paths.Paths == nil {
		primary.Paths.Paths = make(map[string]spec.PathItem)
	}

	m := &merger{
		primary:    primary,
		defs:       make(map[string]spec.Schema),
		paths:      make(map[string]spec.PathItem),
		revertDefs: make(map[string]string),
		replaceMap: make(map[string]string),
	}

	for i := 0; i < len(swaggers); i++ {
		err = m.Add(swaggers[i], scopes[i])
		if err != nil {
			return nil, err
		}
	}

	return m.Swagger(compressLevel)
}

func defaultSwagger() *spec.Swagger {
	return &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:   "Commerce API",
					Version: "1.0",
				},
			},
			Schemes:  []string{"http", "https"},
			Consumes: []string{"application/json"},
			Produces: []string{"application/json"},
			Paths: &spec.Paths{
				Paths: make(map[string]spec.PathItem),
			},
			Definitions: make(map[string]spec.Schema),
		},
	}
}

func (m *merger) Add(swagger *spec.Swagger, scope string) error {
	// + scope
	defs := make(map[string]spec.Schema)
	for k, v := range swagger.Definitions {
		v.ID = scope + k
		defs[scope+k] = v
	}
	swagger.Definitions = defs

	err := replace(swagger, map[string]string{
		prefix: prefix + scope,
	}, "", "")
	if err != nil {
		return err
	}

	// Add paths
	m.AddPaths(swagger.Paths)

	// Add defs
	m.AddDefinitions(swagger)

	return nil
}

func (m *merger) AddPaths(paths *spec.Paths) {
	for k, v := range paths.Paths {
		m.paths[k] = v
	}
}

func (m *merger) AddDefinitions(swagger *spec.Swagger) {
	for _, key := range utils.SortedStringKeys(swagger.Definitions) {
		schema := swagger.Definitions[key]
		uuid := string(toMD5(&schema, swagger))
		if exist, ok := m.revertDefs[uuid]; ok {
			m.replaceMap[key] = exist
		} else {
			m.revertDefs[uuid] = key
			m.defs[key] = schema
		}
	}
}

func (m *merger) Swagger(level int) (*spec.Swagger, error) {
	d := &Dict{}
	for _, k := range utils.SortedStringKeys(m.defs) {
		d.insertStr(k)
	}
	for i := 0; i < level; i++ {
		d.compress()
	}
	shortMap := d.getOrigToShortMap()

	m.primary.Definitions = make(map[string]spec.Schema)
	m.primary.Paths.Paths = make(map[string]spec.PathItem)

	for k, v := range m.defs {
		if short, ok := shortMap[k]; ok {
			k = short
		}
		m.primary.Definitions[k] = v
	}
	for k, v := range m.paths {
		m.primary.Paths.Paths[k] = v
	}
	err := replace(m.primary, m.replaceMap, prefix, suffix)
	if err != nil {
		return nil, err
	}
	err = replace(m.primary, shortMap, prefix, suffix)
	if err != nil {
		return nil, err
	}

	return m.primary, nil
}

// replace content string
func replace(content IMarshaler, replaceMap map[string]string, prefix string, suffix string) error {
	data, err := content.MarshalJSON()
	if err != nil {
		return err
	}
	for from, to := range replaceMap {
		from = prefix + from + suffix
		to = prefix + to + suffix
		data = bytes.Replace(data, []byte(from), []byte(to), -1)
	}
	err = content.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return nil
}

// func toMD5(schema spec.Schema) (string, error) {
// 	hash := md5.New()
// 	data, err := schema.MarshalJSON()
// 	if err != nil {
// 		return "", err
// 	}
// 	hash.Write(data)
// 	checksum := hash.Sum(nil)
// 	return string(checksum), nil
// }

func toMD5(schema *spec.Schema, document interface{}) []byte {
	if schema == nil {
		return nil
	}

	hash := md5.New()

	if utils.IsRef(schema) {
		ref := utils.GetRef(schema, document)
		return toMD5(ref, document)
	}

	if utils.IsArray(schema) {
		return hash.Sum(toMD5(schema.Items.Schema, document))
	}

	if utils.IsObject(schema) {
		hash.Write([]byte("object"))
		for _, k := range utils.SortedStringKeys(schema.Properties) {
			hash.Write([]byte(k))
			child := schema.Properties[k]
			hash.Write(toMD5(&child, document))
		}
		return hash.Sum(nil)
	}

	data, err := schema.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	hash.Write(data)
	return hash.Sum(nil)
}
