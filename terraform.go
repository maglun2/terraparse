package main

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

var rootSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "variable",
			LabelNames: []string{"name"},
		},
		{
			Type:       "locals",
		},
	},
}

var variableBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "description",
		},
		{
			Name: "default",
		},
		{
			Name: "type",
		},
	},
}

type Local struct {
	Name string
	Expr hcl.Expression

	DeclRange hcl.Range
}

type VariableParsingMode rune
const VariableParseLiteral VariableParsingMode = 'L'

type Variable struct {
	Name        string
	Description string
	Default     cty.Value
	Type        cty.Type
	ParsingMode VariableParsingMode

	DescriptionSet bool

	DeclRange hcl.Range
}

func decodeLocalsBlock(block *hcl.Block) ([]*Local) {
	attrs, _ := block.Body.JustAttributes()
	if len(attrs) == 0 {
		return nil
	}

	locals := make([]*Local, 0, len(attrs))
	for name, attr := range attrs {
		locals = append(locals, &Local{
			Name:      name,
			Expr:      attr.Expr,
			DeclRange: attr.Range,
		})
	}
	return locals
}

func decodeVariableBlock(block *hcl.Block, override bool) (*Variable) {
	v := &Variable{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	if !override {
		v.Type = cty.DynamicPseudoType
		v.ParsingMode = VariableParseLiteral
	}

	content, _ := block.Body.Content(variableBlockSchema)

	if attr, exists := content.Attributes["default"]; exists {
		val, _ := attr.Expr.Value(nil)

		if v.Type != cty.NilType {
			var err error
			val, err = convert.Convert(val, v.Type)
			if err != nil {
				val = cty.DynamicVal
			}
		}

		v.Default = val
	}

	return v
}
