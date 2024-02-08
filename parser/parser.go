package parser

import (
	"errors"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/types/typeutil"

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
			obj := obj.(*types.TypeName)
			varType := obj.Type().Underlying()
			switch varType.(type) {
			case *types.Struct:
				methods := typeutil.IntuitiveMethodSet(obj.Type(), nil)
				if len(methods) == 0 {
					continue
				}
				// check if there is a constructor method for the struct
				if !parser.HasConstructorMethod(obj) {
					continue
				}
				class := &models.Class{
					Name: id.Name,
					Path: parser.Package.Path() + "/" + parser.Package.Name(),
					Type: models.ClassDefinition,
				}
				class.Methods = make([]*models.Method, 0)
				parser.Program.Body = append(parser.Program.Body, class)
				for _, method := range methods {
					err := parser.ParseFunction(class, method.Obj().(*types.Func))
					if err != nil {
						return err
					}
				}
				return nil
			}
		}
	}
	return errors.New("no class found")
}

func (parser *Parser) HasConstructorMethod(obj types.Object) bool {
	// check if there is a method in the package that returns the struct type
	for _, object := range parser.Info.Defs {
		if object == nil {
			continue
		}
		switch object.(type) {
		case *types.Func:
			function := object.(*types.Func)
			if function.Type().(*types.Signature).Results().Len() == 1 && function.Name() == "New" {
				if function.Type().(*types.Signature).Results().At(0).Type() == obj.Type() {
					return true
				}
			}
		}
	}
	return false
}

func (parser *Parser) ParseFunction(class *models.Class, function *types.Func) error {
	functionSignature := function.Type().(*types.Signature)
	returnType, err := parser.MapReturnTypeToAstNodeType(functionSignature.Results())
	if err != nil {
		return err
	}
	method := &models.Method{
		Name: function.Name(),
		Type: returnType,
	}
	method.Params = make([]*models.Param, 0)
	for i := 0; i < functionSignature.Params().Len(); i++ {
		param := functionSignature.Params().At(i)
		var optional bool
		switch param.Type().(type) {
		case *types.Pointer:
			optional = true
		}
		method.Params = append(method.Params, &models.Param{
			Name:     param.Name(),
			Type:     parser.MapToAstNodeType(param.Type()),
			Optional: optional,
		})
	}
	class.Methods = append(class.Methods, method)
	return nil
}

func checkIfTypeIsErrorInterface(expr types.Type) bool {
	// construct an error interface
	recvTypeParams := []*types.TypeParam{}
	typeParams := []*types.TypeParam{}
	errorInterface := types.NewInterfaceType([]*types.Func{
		types.NewFunc(0, nil, "Error", types.NewSignatureType(nil, recvTypeParams, typeParams, types.NewTuple(), types.NewTuple(types.NewVar(0, nil, "", types.Typ[types.String])), false)),
	}, nil)
	// check if the type implements the error interface
	switch expr := expr.(type) {
	case *types.Interface:
		return types.Implements(expr, errorInterface)
	case *types.Pointer:
		return types.Implements(expr.Elem(), errorInterface) || types.Implements(expr, errorInterface)
	default:
		return types.Implements(expr, errorInterface)
	}
}

func (parser *Parser) MapReturnTypeToAstNodeType(returnTuple *types.Tuple) (models.AstNode, error) {
	if returnTuple == nil || returnTuple.Len() == 0 || returnTuple.Len() > 2 {
		return nil, errors.New("return tuple should be error or (type, error)")
	}
	if returnTuple.Len() == 1 {
		if checkIfTypeIsErrorInterface(returnTuple.At(0).Type()) {
			return models.BuiltInType{
				Type: models.VoidLiteral,
			}, nil
		}
		return nil, errors.New("return tuple should be error or (type, error)")
	} else if returnTuple.Len() == 2 {
		if !checkIfTypeIsErrorInterface(returnTuple.At(1).Type()) {
			return nil, errors.New("return tuple should be error or (type, error)")
		}
		return parser.MapToAstNodeType(returnTuple.At(0).Type()), nil
	}
	return nil, errors.New("return tuple should be error or (type, error)")
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
			Type: models.CustomNodeLiteral,
			Name: &name,
		}
	case *types.Slice:
		return models.Array{
			Type:    models.ArrayType,
			Generic: parser.MapToAstNodeType(expr.Elem()),
		}
	case *types.Array:
		return models.Array{
			Type:    models.ArrayType,
			Generic: parser.MapToAstNodeType(expr.Elem()),
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
		var optional bool
		switch field.Type().(type) {
		case *types.Pointer:
			optional = true
		}
		properties = append(properties, &models.StructProperty{
			Name:     field.Name(),
			Type:     parser.MapToAstNodeType(field.Type()),
			Optional: optional,
		})
	}
	return models.TypeLiteralStruct{
		Type:       models.TypeLiteral,
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
