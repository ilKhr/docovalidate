package oapi_builder

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/pb33f/libopenapi"
)

type Schemer interface {
	Scheme() string
}

type OapiBuilder struct {
	// created from yamlBuilder in Generate moment
	oapiDoc string

	// combine schemas to one string
	yamlBuilder strings.Builder

	logger *slog.Logger
}

func New(logger *slog.Logger) *OapiBuilder {
	return &OapiBuilder{
		logger: logger,
	}
}

/*
Example:

	openapi: 3.0.0
	info:
		...
	tags:
		...
	servers:
		...
*/
func (ob *OapiBuilder) AddMainInfo(schemer Schemer) *OapiBuilder {

	ob.PasteWithIndent(schemer.Scheme(), 0)

	return ob
}

/*
Example:

	  /api/v1/users:
	    get:
			...
	  /api/v1/users/{id}:
		put:
			...
*/
// AddPaths добавляет пути в спецификацию
// Сюда приходят схемы, в каждой из которых указан путь
// Мы вычисляем, какие пути уже есть и удаляем их из схем
func (ob *OapiBuilder) AddPaths(schemers []Schemer) *OapiBuilder {
	ob.PasteWithIndent("paths:", 0)

	// seen хранит уже добавленные пути и методы
	seen := map[string]map[string]string{} // path -> method -> true

	// Регулярное выражение для поиска endpoint и method
	re := regexp.MustCompile(`(\/\S+):\s*\n\s*(\w+):`)

	for _, s := range schemers {
		cs := s.Scheme()

		matches := re.FindStringSubmatch(cs)

		if len(matches) < 3 {
			continue // пропускаем странные схемы
		}

		path := matches[1]
		method := matches[2]

		// если путь еще не добавлен, то добавляем его
		if _, exists := seen[path]; !exists {
			seen[path] = map[string]string{}
		}

		// если метод уже добавлен, то пропускаем
		if _, exists := seen[path][method]; exists {
			continue
		}

		// удаляем метод и путь из локальной схемы
		// т.к путь идёт перед методом, он также удалится автоматически
		i := strings.Index(cs, method)
		cs = cs[i+len(method)+2:]

		// добавляем схему в общий список
		seen[path][method] = cs

	}

	// проходимся по всему списку и подставляем путь один раз
	for path, methods := range seen {
		ob.PasteWithIndent(path+":", 1)
		for method, cs := range methods {
			ob.PasteWithIndent(method+":", 2)
			ob.PasteWithIndent(cs, 1)
		}
	}

	return ob
}

/*
Example:

	User:
		...
	Error:
		...
*/
func (ob *OapiBuilder) AddComponentsSchemas(schemers []Schemer) *OapiBuilder {

	ob.PasteWithIndent("components:", 0)
	ob.PasteWithIndent("schemas:", 1)

	for _, schemer := range schemers {
		ob.PasteWithIndent(schemer.Scheme(), 2)
	}

	return ob
}

func (ob *OapiBuilder) String() string {
	return ob.yamlBuilder.String()
}

func (ob *OapiBuilder) ValidateAndGenerateYaml(path string) (string, error) {
	var oapiUtil oapiUtil

	if err := oapiUtil.ValidateScheme(ob.yamlBuilder.String()); err != nil {
		log.Println(ob.yamlBuilder.String())

		return "", err
	}

	if err := os.WriteFile(path, []byte(ob.yamlBuilder.String()), 0644); err != nil {
		return "", err
	}

	return ob.yamlBuilder.String(), nil
}

func (ob *OapiBuilder) PasteWithIndent(textWithNewLines string, deepLevel int) *OapiBuilder {
	scanner := bufio.NewScanner(strings.NewReader(textWithNewLines))

	yamlUtil := yamlUtil{}

	for scanner.Scan() {
		line := scanner.Text()

		ob.yamlBuilder.WriteString(yamlUtil.normalizeLine(line, deepLevel))
	}

	return ob
}

func (ob OapiBuilder) GetOapiInBytesFromFile(path string) ([]byte, error) {
	var oapiUtil oapiUtil
	yaml, err := os.ReadFile(path)

	if err := oapiUtil.ValidateScheme(string(yaml)); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	return yaml, nil
}

func (ob OapiBuilder) MustGetOapiInBytesFromFile(path string) []byte {
	yaml, err := ob.GetOapiInBytesFromFile(path)
	if err != nil {
		panic(err)
	}
	return yaml
}

type HandlersWithSchemas struct {
	MainInfoSchemas   Schemer
	PathSchemas       []Schemer
	ComponentsSchemas []Schemer
}

// generateSchemas generate oapi specification and save to file
func (ob *OapiBuilder) generateSchemas(hws HandlersWithSchemas, saveTo string) (*OapiBuilder, error) {
	op := "oapi_builder.GenerateSchemas"

	log := ob.logger.With("op", op)

	log.Info("Start generate oapi specification")

	ob.
		AddMainInfo(hws.MainInfoSchemas).
		AddPaths(hws.PathSchemas).
		AddComponentsSchemas(hws.ComponentsSchemas)

	yaml, err := ob.ValidateAndGenerateYaml(saveTo)

	if err != nil {
		ob.logger.Error("Error validate oapi specification", slog.String("error", err.Error()))
		return nil, err
	}

	ob.oapiDoc = yaml

	log.Info("Oapi specification generated successfully")

	return ob, nil
}

// MustGenerateSchemas generate oapi specification and save to file as GenerateSchemas
//
// Make panic if error
func (ob *OapiBuilder) MustGenerateSchemas(hws HandlersWithSchemas, saveTo string) *OapiBuilder {
	ob, err := ob.generateSchemas(hws, saveTo)
	if err != nil {
		panic(err)
	}
	return ob
}

func (ob *OapiBuilder) GetOapiInBytes() []byte {
	return []byte(ob.oapiDoc)
}

const indent = "\x20\x20"

type yamlUtil struct {
	content strings.Builder
}

func (y *yamlUtil) normalizeLine(line string, indentLevel int) string {
	defer y.content.Reset()

	if line == "" {
		return ""
	}

	line = strings.Trim(line, "\n")
	line = strings.ReplaceAll(line, "\t", indent)

	if indentLevel >= 0 {
		y.content.WriteString(strings.Repeat(indent, indentLevel))
	} else {
		if len(line) > indentLevel {
			line = line[indentLevel*-1:]
		}
	}

	y.content.WriteString(line)

	y.content.WriteString("\n")

	return y.content.String()
}

type oapiUtil struct{}

func (o *oapiUtil) ValidateScheme(yamlContent string) error {
	_, err := libopenapi.NewDocument([]byte(yamlContent))

	if err != nil {
		return fmt.Errorf("невалидная OpenAPI-схема: %w", err)
	}

	return nil
}
