package gen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ettle/strcase"
	"github.com/gamefabric/tfutils/conv/ast"
	"github.com/gamefabric/tfutils/conv/gen"
	"github.com/gamefabric/tfutils/conv/parser"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var typeMap = map[reflect.Type]reflect.Kind{
	reflect.TypeOf(resource.Quantity{}):  reflect.String,
	reflect.TypeOf(intstr.IntOrString{}): reflect.String,
}

type PathGroup struct {
	PrefixPath string
	ObjInfo    []ObjInfo
}

type ObjInfo struct {
	Filename       string
	Obj            any
	Customizations []parser.FieldCustomizerFunc
}

type Generator struct {
	parser *parser.Parser
	gen    *gen.Generator
	namer  gen.Namer

	basePkgPath string
	seen        map[reflect.Type]*ast.ExternalObjectType
}

func New(basePkgPath string) *Generator {
	g := &Generator{
		namer:       &typeNamer{},
		basePkgPath: basePkgPath,
		seen:        map[reflect.Type]*ast.ExternalObjectType{},
	}

	g.parser = parser.New(parser.Options{
		Tag:             "json",
		TypeMap:         typeMap,
		FieldCustomizer: g.customize,
	})
	g.gen = gen.New(gen.Options{
		Namer: g.namer,
	})

	return g
}

func (g *Generator) Generate(basePath string, pkgName string, infos ...ObjInfo) error {
	typs := make([]ast.Type, 0, len(infos))
	for _, info := range infos {
		typ, err := g.parser.Parse(info.Obj, info.Customizations...)
		if err != nil {
			return fmt.Errorf("parsing type %T: %w", info.Obj, err)
		}
		typs = append(typs, typ)

		objTyp, ok := typ.(*ast.ObjectType)
		if !ok {
			return fmt.Errorf("type %T is not an *ast.ObjectType", typ)
		}
		pkgPath := filepath.Join(g.basePkgPath, basePath)
		g.seen[reflect.TypeOf(info.Obj)] = &ast.ExternalObjectType{
			PkgPath:    pkgPath,
			TypeName:   g.namer.Model(objTyp.Name),
			SchemaFunc: g.namer.AttributesFunc(objTyp.Name),
		}
	}

	typeSets, err := ast.ExtractCommonTypes(typs, func(obj *ast.ObjectType) *ast.ExternalObjectType {
		return &ast.ExternalObjectType{
			TypeName:   g.namer.Model(obj.Name),
			SchemaFunc: g.namer.AttributesFunc(obj.Name),
		}
	})
	if err != nil {
		return fmt.Errorf("extracting common types: %w", err)
	}

	for i, t := range typeSets {
		var buf bytes.Buffer
		if err = g.gen.Generate(&buf, pkgName, t...); err != nil {
			return fmt.Errorf("generating code: %w", err)
		}

		path := filepath.Join(basePath, infos[i].Filename)
		if _, err = os.Stat(path); err == nil {
			if err = os.Remove(path); err != nil {
				return fmt.Errorf("removing file %q: %w", path, err)
			}
		}
		if err = os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return fmt.Errorf("creating directory %q: %w", filepath.Base(path), err)
		}
		if err = os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
			return fmt.Errorf("writing file %q: %w", path, err)
		}
	}
	return nil
}

type docable interface {
	Docs() map[string]string
}

type attrable interface {
	Attributes() map[string]string
}

func (g *Generator) customize(ctx parser.Context, field *ast.Field) parser.FieldDecision {
	name := resolveName(ctx.StructField())
	parentObj := ctx.ParentObject()
	if d, ok := parentObj.(docable); ok {
		field.Description = d.Docs()[name]
	}

	attrs := field.Attr
	attrs.SetOptional()
	if a, ok := parentObj.(attrable); ok {
		switch a.Attributes()[name] {
		case "readonly":
			attrs.Computed = true
		case "required":
			attrs.SetRequired()
		}
	}

	if field.Type != nil {
		return parser.FieldDecisionHandled
	}

	typ := ctx.StructField().Type
	if ref, ok := g.seen[typ]; ok {
		field.Type = ref
		return parser.FieldDecisionHandled
	}
	switch typ.Kind() {
	case reflect.Slice:
		ref, ok := g.seen[typ.Elem()]
		if !ok {
			break
		}
		field.Type = &ast.ListType{ElemType: ref}
		return parser.FieldDecisionHandled
	case reflect.Map:
		ref, ok := g.seen[typ.Elem()]
		if !ok {
			break
		}
		field.Type = &ast.MapType{ElemType: ref}
		return parser.FieldDecisionHandled
	}

	if strings.HasSuffix(ctx.Path(), "Template.ObjectMeta") {
		field.Type = &ast.ExternalObjectType{
			PkgPath:    "github.com/gamefabric/terraform-provider-gamefabric/internal/schema/meta",
			TypeName:   "TemplateMetadataModel",
			SchemaFunc: "TemplateMetadataAttributes",
		}
		return parser.FieldDecisionHandled
	}

	switch ctx.StructField().Name {
	case "ObjectMeta":
		field.Type = &ast.ExternalObjectType{
			PkgPath:    "github.com/gamefabric/terraform-provider-gamefabric/internal/schema/meta",
			TypeName:   "MetadataModel",
			SchemaFunc: "MetadataAttributes",
		}
		return parser.FieldDecisionHandled
	case "TypeMeta", "Status":
		return parser.FieldDecisionSkip
	}
	return parser.FieldDecisionNoOpinion
}

func resolveName(sf reflect.StructField) string {
	name := sf.Tag.Get("json")
	if tagName, _, ok := strings.Cut(name, ","); ok {
		name = tagName
	}
	if name == "" || name == "-" {
		return ""
	}
	return name
}

type typeNamer struct{}

func (n typeNamer) Model(name string) string { return strcase.ToGoPascal(name) + "Model" }
func (n typeNamer) AttributesFunc(name string) string {
	return strcase.ToGoPascal(name) + "Attributes"
}
