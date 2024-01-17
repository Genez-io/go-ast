package models

type Program struct {
	Body []AstNode `json:"body"`
}

type AstNodeType string

const (
	ClassDefinition     AstNodeType = "ClassDefinition"
	MethodDefinttion    AstNodeType = "MethodDefinttion"
	ParameterDefinition AstNodeType = "ParameterDefinition"
	StringLiteral       AstNodeType = "StringLiteral"
	IntLiteral          AstNodeType = "IntLiteral"
	BoolLiteral         AstNodeType = "BoolLiteral"
	FloatLiteral        AstNodeType = "FloatLiteral"
	VoidLiteral         AstNodeType = "VoidLiteral"
	AnyLiteral          AstNodeType = "AnyLiteral"
	StructLiteral       AstNodeType = "StructLiteral"
	TypeLiteral         AstNodeType = "TypeLiteral"
	TypeAlias           AstNodeType = "TypeAlias"
	CustomTypeLiteral   AstNodeType = "CustomTypeLiteral"
)

type AstNode interface {
	GetName() *string
	GetType() AstNodeType
	GetPAth() *string
}

type BuiltInType struct {
	Type AstNodeType `json:"type"`
}

func (builtInType BuiltInType) GetName() *string {
	name := string(builtInType.Type)
	return &name
}

func (builtInType BuiltInType) GetType() AstNodeType {
	return builtInType.Type
}

func (builtInType BuiltInType) GetPAth() *string {
	return nil
}

type CustomType struct {
	Name *string     `json:"name"`
	Type AstNodeType `json:"type"`
}

func (customType CustomType) GetName() *string {
	return customType.Name
}

func (customType CustomType) GetType() AstNodeType {
	return CustomTypeLiteral
}

func (customType CustomType) GetPAth() *string {
	return nil
}

type Class struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Methods []*Method `json:"methods"`
}

func (class Class) GetName() *string {
	return &class.Name
}

func (class Class) GetType() AstNodeType {
	return ClassDefinition
}

func (class Class) GetPAth() *string {
	return &class.Path
}

type Method struct {
	Name   string   `json:"name"`
	Params []*Param `json:"params"`
	Type   AstNode  `json:"returnType"`
}

type Param struct {
	Name     string  `json:"name"`
	Type     AstNode `json:"paramType"`
	Optional bool    `json:"optional"`
}

type StructProperty struct {
	Name     string  `json:"name"`
	Type     AstNode `json:"type"`
	Optional bool    `json:"optional"`
}

type TypeLiteralStruct struct {
	Properties []*StructProperty `json:"properties"`
}

func (typeLiteralStruct TypeLiteralStruct) GetName() *string {
	return nil
}

func (typeLiteralStruct TypeLiteralStruct) GetType() AstNodeType {
	return TypeLiteral
}

func (typeLiteralStruct TypeLiteralStruct) GetPAth() *string {
	return nil
}

type Struct struct {
	Name        string             `json:"name"`
	TypeLiteral *TypeLiteralStruct `json:"typeLiteral"`
	Path        string             `json:"path"`
	Type        AstNodeType        `json:"type"`
}

func (structLiteral Struct) GetName() *string {
	return &structLiteral.Name
}

func (structLiteral Struct) GetType() AstNodeType {
	return StructLiteral
}

func (structLiteral Struct) GetPAth() *string {
	return &structLiteral.Path
}