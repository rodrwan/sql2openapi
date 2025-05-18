package generator

import (
	"fmt"
	"strings"

	"github.com/rodrwan/sql2openapi/internal/config"
	"github.com/rodrwan/sql2openapi/internal/parser"

	"gopkg.in/yaml.v2"
)

type OpenAPI struct {
	OpenAPI    string                `yaml:"openapi"`
	Info       Info                  `yaml:"info"`
	Paths      map[string]PathItem   `yaml:"paths"`
	Components Components            `yaml:"components"`
	Security   []map[string][]string `yaml:"security,omitempty"` // si quieres seguridad global opcional
}

type Info struct {
	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

type Components struct {
	Schemas         map[string]Schema         `yaml:"schemas"`
	SecuritySchemes map[string]SecurityScheme `yaml:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type         string `yaml:"type"`
	Scheme       string `yaml:"scheme"`
	BearerFormat string `yaml:"bearerFormat,omitempty"`
}

type Schema struct {
	Type       string              `yaml:"type"`
	Properties map[string]Property `yaml:"properties"`
	Required   []string            `yaml:"required,omitempty"`
}

type Property struct {
	Type        string `yaml:"type"`
	Format      string `yaml:"format,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type PathItem struct {
	Get    *Operation `yaml:"get,omitempty"`
	Post   *Operation `yaml:"post,omitempty"`
	Put    *Operation `yaml:"put,omitempty"`
	Delete *Operation `yaml:"delete,omitempty"`
}

type Operation struct {
	Summary     string                `yaml:"summary"`
	Parameters  []Parameter           `yaml:"parameters,omitempty"`
	RequestBody *RequestBody          `yaml:"requestBody,omitempty"`
	Responses   map[string]Response   `yaml:"responses"`
	Security    []map[string][]string `yaml:"security,omitempty"`
}

type Response struct {
	Description string             `yaml:"description"`
	Content     map[string]Content `yaml:"content,omitempty"`
}

type RequestBody struct {
	Content map[string]Content `yaml:"content"`
}

type Content struct {
	Schema SchemaRef `yaml:"schema"`
}

type SchemaRef struct {
	Ref        string              `yaml:"$ref,omitempty"`
	Type       string              `yaml:"type,omitempty"`
	Items      *SchemaRef          `yaml:"items,omitempty"`
	Properties map[string]Property `yaml:"properties,omitempty"`
	Required   []string
}

// ------------------------------------------------------

func GenerateOpenAPI(tables []parser.Table, cfg *config.Config) ([]byte, error) {
	openapi := OpenAPI{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:   "Generated API",
			Version: "1.0.0",
		},
		Paths:      make(map[string]PathItem),
		Components: Components{Schemas: make(map[string]Schema)},
	}

	if cfg != nil && cfg.Security.Scheme == "bearer" {
		openapi.Components.SecuritySchemes = map[string]SecurityScheme{
			"bearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: cfg.Security.Format,
			},
		}
	}
	security := []map[string][]string{
		{"bearerAuth": {}},
	}

	for _, table := range tables {
		schema := Schema{
			Type:       "object",
			Properties: make(map[string]Property),
		}
		var required []string

		for _, col := range table.Columns {
			if col.ForeignKey != nil {
				parent := col.ForeignKey.RefTable
				child := table.Name
				path := fmt.Sprintf("/%s/{id}/%s", parent, child)

				openapi.Paths[path] = PathItem{
					Get: &Operation{
						Summary:   fmt.Sprintf("List %s by %s ID", child, parent),
						Responses: defaultArrayResponse(child),
					},
				}
			}

			prop := Property{
				Type: sqlToOpenAPIType(col.Type),
			}
			if prop.Type == "integer" {
				prop.Format = "int32"
			}
			if col.ForeignKey != nil {
				prop.Description = fmt.Sprintf("Foreign key to %s.%s", col.ForeignKey.RefTable, col.ForeignKey.RefCol)
			}

			schema.Properties[col.Name] = prop
			if col.Primary {
				required = append(required, col.Name)
			}
			if col.Required {
				required = append(required, col.Name)
			}
		}
		if len(required) > 0 {
			schema.Required = required
		}

		openapi.Components.Schemas[table.Name] = schema
		pkType := getPrimaryKeyType(table)

		item := PathItem{
			Get: &Operation{
				Summary:   fmt.Sprintf("Get all %s", table.Name),
				Responses: defaultResponse(table.Name),
			},
			Post: &Operation{
				Summary:     fmt.Sprintf("Create new %s", table.Name),
				RequestBody: requestBody(table.Name),
				Responses:   defaultResponse(table.Name),
				Security: []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		}

		path := fmt.Sprintf("/%s", table.Name)
		pathWithID := fmt.Sprintf("/%s/{id}", table.Name)
		itemWithID := PathItem{
			Get: &Operation{
				Summary:    fmt.Sprintf("Get %s by ID", table.Name),
				Parameters: pathParamID(pkType),
				Responses:  defaultResponse(table.Name),
			},
			Put: &Operation{
				Summary:     fmt.Sprintf("Update %s by ID", table.Name),
				Parameters:  pathParamID(pkType),
				RequestBody: requestBody(table.Name),
				Responses:   defaultResponse(table.Name),
			},
			Delete: &Operation{
				Summary:    fmt.Sprintf("Delete %s by ID", table.Name),
				Parameters: pathParamID(pkType),
				Responses:  noContentResponse(),
			},
		}

		if item.Get != nil && isProtected(path, "get", cfg) {
			item.Get.Security = security
		}
		if item.Post != nil && isProtected(path, "post", cfg) {
			item.Post.Security = security
		}
		if item.Put != nil && isProtected(path, "put", cfg) {
			item.Put.Security = security
		}
		if item.Delete != nil && isProtected(path, "delete", cfg) {
			item.Delete.Security = security
		}

		if itemWithID.Get != nil && isProtected(pathWithID, "get", cfg) {
			itemWithID.Get.Security = security
		}
		if itemWithID.Post != nil && isProtected(pathWithID, "post", cfg) {
			itemWithID.Post.Security = security
		}
		if itemWithID.Put != nil && isProtected(pathWithID, "put", cfg) {
			itemWithID.Put.Security = security
		}
		if itemWithID.Delete != nil && isProtected(pathWithID, "delete", cfg) {
			itemWithID.Delete.Security = security
		}

		openapi.Paths[path] = item
		openapi.Paths[pathWithID] = itemWithID
	}

	if cfg != nil {
		for _, ep := range cfg.CustomEndpoints {
			path := ep.Path
			if openapi.Paths[path].Get != nil ||
				openapi.Paths[path].Post != nil {
				continue // ya existe algo
			}

			item := PathItem{}
			op := &Operation{
				Summary: ep.Summary,
				Responses: map[string]Response{
					"200": {
						Description: "OK",
						Content: map[string]Content{
							"application/json": {
								Schema: SchemaRefFromMap(ep.ResponseSchema),
							},
						},
					},
				},
			}

			if ep.RequestSchema != nil {
				op.RequestBody = &RequestBody{
					Content: map[string]Content{
						"application/json": {
							Schema: SchemaRefFromMap(ep.RequestSchema),
						},
					},
				}
			}

			switch strings.ToLower(ep.Method) {
			case "post":
				item.Post = op
			case "get":
				item.Get = op
			case "put":
				item.Put = op
			case "delete":
				item.Delete = op
			}

			openapi.Paths[path] = item
		}
	}

	return yaml.Marshal(openapi)
}

// ------------------------------------------------------

func sqlToOpenAPIType(sqlType string) string {
	switch strings.ToUpper(sqlType) {
	case "INT", "INTEGER", "SERIAL":
		return "integer"
	case "TEXT", "VARCHAR", "UUID":
		return "string"
	default:
		return "string"
	}
}

func defaultResponse(schemaRef string) map[string]Response {
	return map[string]Response{
		"200": {
			Description: "OK",
			Content: map[string]Content{
				"application/json": {
					Schema: SchemaRef{
						Ref: "#/components/schemas/" + schemaRef,
					},
				},
			},
		},
	}
}

func requestBody(schemaRef string) *RequestBody {
	return &RequestBody{
		Content: map[string]Content{
			"application/json": {
				Schema: SchemaRef{
					Ref: "#/components/schemas/" + schemaRef,
				},
			},
		},
	}
}

func defaultArrayResponse(schemaRef string) map[string]Response {
	return map[string]Response{
		"200": {
			Description: "OK",
			Content: map[string]Content{
				"application/json": {
					Schema: SchemaRef{
						Type: "array",
						Items: &SchemaRef{
							Ref: "#/components/schemas/" + schemaRef,
						},
					},
				},
			},
		},
	}
}

func noContentResponse() map[string]Response {
	return map[string]Response{
		"204": {
			Description: "No Content",
		},
	}
}

func pathParamID(typeStr string) []Parameter {
	return []Parameter{
		{
			Name:     "id",
			In:       "path",
			Required: true,
			Schema: map[string]string{
				"type": typeStr,
			},
		},
	}
}

type Parameter struct {
	Name     string            `yaml:"name"`
	In       string            `yaml:"in"`
	Required bool              `yaml:"required"`
	Schema   map[string]string `yaml:"schema"`
}

func getPrimaryKeyType(table parser.Table) string {
	for _, col := range table.Columns {
		if col.Primary {
			return sqlToOpenAPIType(col.Type)
		}
	}
	return "string" // fallback
}

func isProtected(path, method string, cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	for _, p := range cfg.Security.Protected {
		if p.Path == path {
			for _, m := range p.Methods {
				if strings.EqualFold(m, method) {
					return true
				}
			}
		}
	}
	return false
}

func SchemaRefFromMap(m map[string]interface{}) SchemaRef {
	// Si contiene $ref, devolvemos solo eso
	if ref, ok := m["$ref"].(string); ok {
		return SchemaRef{Ref: ref}
	}

	schema := SchemaRef{}

	if t, ok := m["type"].(string); ok {
		schema.Type = t
	}

	// Procesar "properties"
	if props, ok := m["properties"].(map[interface{}]interface{}); ok {
		schema.Properties = map[string]Property{}
		for k, v := range props {
			propName := fmt.Sprintf("%v", k)
			propData := v.(map[interface{}]interface{})

			prop := Property{}
			if typeVal, ok := propData["type"].(string); ok {
				prop.Type = typeVal
			}
			schema.Properties[propName] = prop
		}
	}

	// Procesar "required"
	if reqs, ok := m["required"].([]interface{}); ok {
		for _, r := range reqs {
			schema.Required = append(schema.Required, fmt.Sprintf("%v", r))
		}
	}

	return schema
}
