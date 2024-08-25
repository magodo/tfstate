package tfstate

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestUnmarshalToCty(t *testing.T) {
	cases := []struct {
		name   string
		obj    map[string]interface{}
		typ    cty.Type
		expect cty.Value
	}{
		{
			name: "basic",
			obj: map[string]interface{}{
				"bool":   true,
				"string": "abc",
				"number": float64(123),
				"list": []interface{}{
					float64(1),
					float64(2),
					float64(3),
				},
				"set": []interface{}{
					float64(1),
					float64(2),
					float64(3),
				},
				"tuple": []interface{}{true, float64(1), "a"},
				"object": map[string]interface{}{
					"bool": true,
				},
				"nested_list": []interface{}{
					map[string]interface{}{
						"name": "a",
					},
					map[string]interface{}{
						"name": "b",
					},
				},
				"nested_set": []interface{}{
					map[string]interface{}{
						"name": "a",
					},
					map[string]interface{}{
						"name": "b",
					},
				},
				"nested_tuple": []interface{}{
					"a",
					map[string]interface{}{
						"name": "a",
					},
				},
				"dynamic_bool":   true,
				"dynamic_number": float64(123),
				"dynamic_string": "abc",
				"dynamic_tuple": []interface{}{
					float64(1),
					float64(2),
					float64(3),
				},
				"dynamic_object": map[string]interface{}{
					"name": "a",
				},
			},
			typ: cty.Object(map[string]cty.Type{
				"bool":           cty.Bool,
				"string":         cty.String,
				"number":         cty.Number,
				"list":           cty.List(cty.Number),
				"set":            cty.Set(cty.Number),
				"tuple":          cty.Tuple([]cty.Type{cty.Bool, cty.Number, cty.String}),
				"object":         cty.Object(map[string]cty.Type{"bool": cty.Bool}),
				"nested_list":    cty.List(cty.Object(map[string]cty.Type{"name": cty.String})),
				"nested_set":     cty.Set(cty.Object(map[string]cty.Type{"name": cty.String})),
				"nested_tuple":   cty.Tuple([]cty.Type{cty.String, cty.Object(map[string]cty.Type{"name": cty.String})}),
				"dynamic_bool":   cty.DynamicPseudoType,
				"dynamic_number": cty.DynamicPseudoType,
				"dynamic_string": cty.DynamicPseudoType,
				"dynamic_tuple":  cty.DynamicPseudoType,
				"dynamic_object": cty.DynamicPseudoType,
			}),
			expect: cty.ObjectVal(map[string]cty.Value{
				"bool":   cty.BoolVal(true),
				"string": cty.StringVal("abc"),
				"number": cty.NumberFloatVal(123),
				"list": cty.ListVal([]cty.Value{
					cty.NumberFloatVal(1),
					cty.NumberFloatVal(2),
					cty.NumberFloatVal(3),
				}),
				"set": cty.SetVal([]cty.Value{
					cty.NumberFloatVal(1),
					cty.NumberFloatVal(2),
					cty.NumberFloatVal(3),
				}),
				"tuple": cty.TupleVal([]cty.Value{
					cty.BoolVal(true),
					cty.NumberFloatVal(1),
					cty.StringVal("a"),
				}),
				"object": cty.ObjectVal(map[string]cty.Value{"bool": cty.BoolVal(true)}),
				"nested_list": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("a")}),
					cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("b")}),
				}),
				"nested_set": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("a")}),
					cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("b")}),
				}),
				"nested_tuple": cty.TupleVal([]cty.Value{
					cty.StringVal("a"),
					cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("a")}),
				}),
				"dynamic_bool":   cty.BoolVal(true),
				"dynamic_number": cty.NumberFloatVal(123),
				"dynamic_string": cty.StringVal("abc"),
				"dynamic_tuple": cty.TupleVal([]cty.Value{
					cty.NumberFloatVal(1),
					cty.NumberFloatVal(2),
					cty.NumberFloatVal(3),
				}),
				"dynamic_object": cty.ObjectVal(map[string]cty.Value{
					"name": cty.StringVal("a"),
				}),
			}),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			v, err := UnmarshalToCty(tt.obj, tt.typ)
			require.NoError(t, err)
			require.Equal(t, tt.expect, v)
		})
	}
}
