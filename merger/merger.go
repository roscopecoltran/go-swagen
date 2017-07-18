package merger

import (
	"bytes"
	"crypto/md5"
	"errors"

	"github.com/go-openapi/spec"
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
	err := m.AddPaths(swagger.Paths)
	if err != nil {
		return err
	}

	err = m.AddDefinitions(swagger.Definitions, scope)
	if err != nil {
		return err
	}

	return nil
}

func (m *merger) AddPaths(paths *spec.Paths) error {
	for k, v := range paths.Paths {
		m.paths[k] = v
	}

	return nil
}

func (m *merger) AddDefinitions(defs spec.Definitions, scope string) error {
	for key, schema := range defs {
		uuid, err := toMD5(schema)
		if err != nil {
			return err
		}
		if exist, ok := m.revertDefs[uuid]; ok {
			m.replaceMap[key] = exist
		} else {
			scopedKey := scope + key
			m.revertDefs[uuid] = scopedKey
			m.defs[scopedKey] = schema
			m.replaceMap[key] = scopedKey
		}
	}

	return nil
}

func (m *merger) Swagger(level int) (*spec.Swagger, error) {
	d := &Dict{}
	for k := range m.defs {
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
	err := replace(m.primary, m.replaceMap)
	if err != nil {
		return nil, err
	}
	err = replace(m.primary, shortMap)
	if err != nil {
		return nil, err
	}

	return m.primary, nil
}

// replace content string
func replace(content IMarshaler, replaceMap map[string]string) error {
	data, err := content.MarshalJSON()
	if err != nil {
		return err
	}

	for from, to := range replaceMap {
		data = bytes.Replace(data, []byte("#/definitions/"+from+"\""), []byte("#/definitions/"+to+"\""), -1)
	}
	err = content.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return nil
}

func toMD5(schema spec.Schema) (string, error) {
	hash := md5.New()
	data, err := schema.MarshalJSON()
	if err != nil {
		return "", err
	}
	hash.Write(data)
	checksum := hash.Sum(nil)
	return string(checksum), nil
}
