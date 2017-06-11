package merger

import (
	"bytes"
	"crypto/md5"
	"errors"
	"path"
	"runtime"

	"github.com/go-openapi/spec"
	"github.com/xreception/go-swagen/utils"
)

// IMarshaler interface
type IMarshaler interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

type merger struct {
	swagger       *spec.Swagger
	compressLevel int
	revertDefs    map[string]string
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
		primary, err = defaultSwagger()
		if err != nil {
			return nil, err
		}
	}
	if primary.Paths.Paths == nil {
		primary.Paths.Paths = make(map[string]spec.PathItem)
	}

	m := &merger{
		swagger:       primary,
		compressLevel: compressLevel,
		revertDefs:    make(map[string]string),
	}

	for i := 0; i < len(swaggers); i++ {
		err = m.Add(scopes[i], swaggers[i])
		if err != nil {
			return nil, err
		}
	}

	err = m.Compress(compressLevel)
	if err != nil {
		return nil, err
	}

	return m.swagger, nil
}

func defaultSwagger() (*spec.Swagger, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return nil, errors.New("could not open default swagger.json file")
	}

	filepath := path.Join(path.Dir(filename), "./templates/swagger.json")
	swagger, err := utils.LoadSpec(filepath)
	if err != nil {
		return nil, err
	}

	return swagger, nil
}

func (m *merger) Add(scope string, swagger *spec.Swagger) error {
	// if len(scope) == 0 {
	// 	return errors.New("Must have a scope for swagger to be merged")
	// }

	err := m.Normalizr(scope, swagger)
	if err != nil {
		return err
	}

	err = m.AddPaths(swagger.Paths)
	if err != nil {
		return err
	}

	err = m.AddDefinitions(swagger.Definitions)
	if err != nil {
		return err
	}

	return nil
}

func (m *merger) Normalizr(scope string, swagger *spec.Swagger) error {
	// Put scope in paths' definition ref
	from := "#/definitions/"
	to := "#/definitions/" + scope
	err := replace(swagger, map[string]string{from: to})
	if err != nil {
		return err
	}

	// definitions
	replaceMap := make(map[string]string)
	defs := make(map[string]spec.Schema)
	hash := md5.New()

	for k, schema := range swagger.Definitions {
		key := scope + k
		data, err := schema.MarshalJSON()
		if err != nil {
			return err
		}
		checksum := hash.Sum(data)
		uuid := string(checksum)
		if exist, ok := m.revertDefs[uuid]; ok {
			replaceMap[key] = exist
		} else {
			m.revertDefs[uuid] = key
			defs[key] = schema
		}
	}

	swagger.Definitions = defs

	err = replace(swagger, replaceMap)
	if err != nil {
		return err
	}

	return nil
}

func (m *merger) AddPaths(paths *spec.Paths) error {

	for k, v := range paths.Paths {
		m.swagger.Paths.Paths[k] = v
	}

	return nil
}

func (m *merger) AddDefinitions(defs spec.Definitions) error {
	for k, v := range defs {
		m.swagger.Definitions[k] = v
	}

	return nil
}

func (m *merger) Compress(level int) error {
	d := &Dict{}
	for k := range m.swagger.Definitions {
		d.insertStr(k)
	}
	for i := 0; i < level; i++ {
		d.compress()
	}
	replaceMap := d.getOrigToShortMap()
	return replace(m.swagger, replaceMap)
}

// replace content string
func replace(content IMarshaler, replaceMap map[string]string) error {
	data, err := content.MarshalJSON()
	if err != nil {
		return err
	}

	for from, to := range replaceMap {
		data = bytes.Replace(data, []byte(from), []byte(to), -1)
	}
	err = content.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return nil
}
