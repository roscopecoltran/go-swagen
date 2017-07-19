package reactReduxTypescript

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"io"

	"github.com/go-openapi/swag"
	"github.com/xreception/go-swagen/utils"
)

var templates *Repository

// FuncMap is a map with default functions for use n the templates.
// These are available in every template
var FuncMap template.FuncMap = map[string]interface{}{
	"CamelCase":     utils.CamelCase,
	"InterfaceCase": utils.InterfaceCase,
	"PluralCase":    utils.PluralCase,
}

func init() {
	templates = NewRepository(FuncMap)
	templates.LoadDefaults()
}

var assets = map[string][]byte{
	"action.tmpl":   MustAsset("templates/action.tmpl"),
	"api.tmpl":      MustAsset("templates/api.tmpl"),
	"constant.tmpl": MustAsset("templates/constant.tmpl"),
	"schema.tmpl":   MustAsset("templates/schema.tmpl"),
}

// NewRepository creates a new template repository with the provided functions defined
func NewRepository(funcs template.FuncMap) *Repository {
	repo := Repository{
		files:     make(map[string]string),
		templates: make(map[string]*template.Template),
		funcs:     funcs,
	}

	if repo.funcs == nil {
		repo.funcs = make(template.FuncMap)
	}

	return &repo
}

// Repository is the repository for the generator templates.
type Repository struct {
	files     map[string]string
	templates map[string]*template.Template
	funcs     template.FuncMap
}

// LoadDefaults will load the embedded templates
func (t *Repository) LoadDefaults() {

	for name, asset := range assets {
		if err := t.addFile(name, string(asset)); err != nil {
			log.Fatal(err)
		}
	}
}

func (t *Repository) addFile(name, data string) error {
	fileName := name
	name = swag.ToJSONName(strings.TrimSuffix(name, ".tmpl"))

	templ, err := template.New(name).Funcs(t.funcs).Parse(data)

	if err != nil {
		return fmt.Errorf("Failed to load template %s: %v", name, err)
	}

	// Add each defined tempalte into the cache
	for _, template := range templ.Templates() {

		t.files[template.Name()] = fileName
		t.templates[template.Name()] = template.Lookup(template.Name())
	}

	return nil
}

// DumpTemplates prints out a dump of all the defined templates, where they are defined and what their dependencies are.
func (t *Repository) DumpTemplates() {
	fmt.Println("# Templates")
	for name := range t.templates {
		fmt.Printf("## %s defined in `%s`\n", name, t.files[name])
	}
}

// ExecuteTemplate generates file with template
func (t *Repository) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	tmpl := t.templates[name]
	return tmpl.Execute(wr, data)
}
