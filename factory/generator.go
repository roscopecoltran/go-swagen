package factory

import (
	"fmt"

	"github.com/go-openapi/spec"
)

// factories stores an internal mapping between generator names and their respective factories.
var factories = make(map[string]IGeneratorFactory)

// IGeneratorFactory is an interface of factory which creates generator.
type IGeneratorFactory interface {
	// Create returns a new Generator with the given parameters.
	// Parameters will vary by generator and may be ignored.
	// Each parameter key must only consist of lowercase letters and numbers.
	Create(parameters map[string]interface{}) (IGenerator, error)
}

// IGenerator is created by Factory
type IGenerator interface {
	// Parse generate code in out folder.
	// The input is spec.Swagger
	Parse(swagger *spec.Swagger, out string) error

	// ParseFile generate code in out folder.
	// The input is the path of swagger.json file
	ParseFile(in string, out string) error
}

// Register makes a factory available by the provided name.
// If Register is called twice with the same name or if factory is nil, it panics.
// Additionally, it is not concurrency safe.
// Call it in init() of each concrete generator.
func Register(name string, factory IGeneratorFactory) {
	if factory == nil {
		panic("Must not provide nil Factory")
	}
	_, registered := factories[name]
	if registered {
		panic(fmt.Sprintf("Factory named %s already registered", name))
	}
	factories[name] = factory
}

// Create a new Generator with the given name and parameters.
// To use a generator, the Factory must first be registered with the given name.
// If no drivers are found, an InvalidStorageDriverError is returned.
func Create(name string, parameters map[string]interface{}) (IGenerator, error) {
	factory, ok := factories[name]
	if !ok {
		return nil, InvalidGeneratorError{name}
	}
	return factory.Create(parameters)
}

// InvalidGeneratorError records an attempt to construct an unregistered generator.
type InvalidGeneratorError struct {
	Name string
}

func (err InvalidGeneratorError) Error() string {
	return fmt.Sprintf("Generator not registered: %s", err.Name)
}
