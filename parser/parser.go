package parser

import (
	"go/ast"
	"go/types"

	"gnz-go-ast/models"
)

type Parser struct {
	Program *models.Program
	Info    *types.Info
	Package *types.Package
}

func New(info *types.Info, pkg *types.Package) *Parser {
	return &Parser{
		Program: &models.Program{},
		Info:    info,
		Package: pkg,
	}
}

func (parser *Parser) Parse(file *ast.File) error {
	parser.Program.Body = make([]models.AstNode, 0)
	for id, obj := range parser.Info.Defs {
		if obj == nil {
			continue
		}
		switch obj.(type) {
		case *types.TypeName:
			varType := obj.Type().Underlying()
			switch varType.(type) {
			case *types.Struct:
				methods := types.NewMethodSet(obj.Type())
				if methods.Len() == 0 {
					continue
				}
				class := &models.Class{
					Name: id.Name,
					Path: parser.Package.Path() + "/" + parser.Package.Name(),
				}
				class.Methods = make([]*models.Method, 0)
				parser.Program.Body = append(parser.Program.Body, class)
				for i := 0; i < methods.Len(); i++ {
					method := methods.At(i)
					parser.ParseFunction(class, method.Obj().(*types.Func))
				}
			}
		}
	}
	return nil
}

func (parser *Parser) ParseFunction(class *models.Class, function *types.Func) error {
	functionSignature := function.Type().(*types.Signature)
	method := &models.Method{
		Name: function.Name(),
		Type: parser.MapReturnTypeToAstNodeType(functionSignature.Results()),
	}
	method.Params = make([]*models.Param, 0)
	for i := 0; i < functionSignature.Params().Len(); i++ {
		param := functionSignature.Params().At(i)
		method.Params = append(method.Params, &models.Param{
			Name: param.Name(),
			Type: parser.MapToAstNodeType(param.Type()),
		})
	}
	class.Methods = append(class.Methods, method)
	return nil
}

func (parser *Parser) MapReturnTypeToAstNodeType(returnTuple *types.Tuple) models.AstNode {
	if returnTuple == nil {
		return models.BuiltInType{
			Type: models.VoidLiteral,
		}
	}
	if returnTuple.Len() == 0 {
		return models.BuiltInType{
			Type: models.VoidLiteral,
		}
	}
	return parser.MapToAstNodeType(returnTuple.At(0).Type())
}

func (parser *Parser) MapToAstNodeType(expr types.Type) models.AstNode {
	switch expr := expr.(type) {
	case *types.Basic:
		switch expr.Kind() {
		case types.String:
			return models.BuiltInType{
				Type: models.StringLiteral,
			}
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
			return models.BuiltInType{
				Type: models.IntLiteral,
			}
		case types.Bool:
			return models.BuiltInType{
				Type: models.BoolLiteral,
			}
		case types.Float32, types.Float64:
			return models.BuiltInType{
				Type: models.FloatLiteral,
			}
		}
	case *types.Struct:
		return parser.ParseStruct(expr)
	case *types.Pointer:
		return parser.MapToAstNodeType(expr.Elem())
	case *types.Named:
		name := expr.Obj().Name()
		path := expr.Obj().Pkg().Path() + "/" + expr.Obj().Pkg().Name()
		parser.AddTypeDefinition(name, path, expr.Underlying())
		return models.CustomType{
			Name: &name,
		}
	}
	return models.BuiltInType{
		Type: models.AnyLiteral,
	}
}

func (parser *Parser) ParseStruct(expr *types.Struct) models.TypeLiteralStruct {
	properties := make([]*models.StructProperty, 0)
	for i := 0; i < expr.NumFields(); i++ {
		field := expr.Field(i)
		properties = append(properties, &models.StructProperty{
			Name: field.Name(),
			Type: parser.MapToAstNodeType(field.Type()),
		})
	}
	return models.TypeLiteralStruct{
		Properties: properties,
	}
}

func (parser *Parser) AddTypeDefinition(name string, path string, definedType types.Type) {
	for _, node := range parser.Program.Body {
		if *node.GetName() == name {
			return
		}
	}
	switch definedType.(type) {
	case *types.Struct:
		structType := definedType.(*types.Struct)
		typeLiteral := parser.ParseStruct(structType)
		parser.Program.Body = append(parser.Program.Body, &models.Struct{
			Name:        name,
			TypeLiteral: &typeLiteral,
			Type:        models.StructLiteral,
			Path:        path,
		})
	}
}
