package tfstate_test

import (
	"encoding/json"
	"math/big"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/magodo/tfstate"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func TestFromJSONStateResource(t *testing.T) {
	state := &tfjson.StateResource{
		Address:      "demo_resource_foo.test",
		Mode:         tfjson.ManagedResourceMode,
		Type:         "demo_resource_foo",
		Name:         "test",
		Index:        1,
		ProviderName: "registry.terraform.io/magodo/demo",
		AttributeValues: map[string]interface{}{
			"attr_str":    "some string",
			"attr_int":    -1,
			"attr_uint":   1,
			"attr_float":  0.1,
			"attr_number": 0.5,
			"attr_bool":   true,
			"attr_list":   []int{1, 2, 3},
			"attr_set":    []int{1, 2, 3},
			"attr_map": map[string]interface{}{
				"key": "value",
			},
			"attr_tuple": []interface{}{
				1,
				map[string]interface{}{"foo": "bar"},
				[]int{1, 2, 3},
			},
			"object": map[string]interface{}{
				"field": 1,
				"nest": map[string]interface{}{
					"field": "a",
				},
			},
		},
		SensitiveValues: json.RawMessage{1, 2, 3},
		DependsOn:       []string{"dep"},
		Tainted:         true,
		DeposedKey:      "key",
	}
	schemas := &tfjson.ProviderSchemas{
		Schemas: map[string]*tfjson.ProviderSchema{
			"registry.terraform.io/magodo/demo": {
				ResourceSchemas: map[string]*tfjson.Schema{
					"demo_resource_foo": {
						Block: &tfjson.SchemaBlock{
							Attributes: map[string]*tfjson.SchemaAttribute{
								"attr_str": {
									AttributeType: cty.String,
								},
								"attr_int": {
									AttributeType: cty.Number,
								},
								"attr_uint": {
									AttributeType: cty.Number,
								},
								"attr_float": {
									AttributeType: cty.Number,
								},
								"attr_number": {
									AttributeType: cty.Number,
								},
								"attr_bool": {
									AttributeType: cty.Bool,
								},
								"attr_list": {
									AttributeType: cty.List(cty.Number),
								},
								"attr_set": {
									AttributeType: cty.Set(cty.Number),
								},
								"attr_map": {
									AttributeType: cty.Map(cty.String),
								},
								"attr_tuple": {
									AttributeType: cty.Tuple([]cty.Type{
										cty.Number,
										cty.Map(cty.String),
										cty.List(cty.Number),
									}),
								},
								"object": {
									AttributeType: cty.Object(map[string]cty.Type{
										"field": cty.Number,
										"nest": cty.Object(map[string]cty.Type{
											"field": cty.String,
										}),
									}),
								},
							},
						},
					},
				},
			},
		},
	}
	expectResourceWithoutValue := &tfstate.StateResource{
		Address:         "demo_resource_foo.test",
		Mode:            tfjson.ManagedResourceMode,
		Type:            "demo_resource_foo",
		Name:            "test",
		Index:           1,
		ProviderName:    "registry.terraform.io/magodo/demo",
		Value:           cty.NilVal, // This is tested separately
		SensitiveValues: json.RawMessage{1, 2, 3},
		DependsOn:       []string{"dep"},
		Tainted:         true,
		DeposedKey:      "key",
	}

	// We are checking the cty value via comparing the Go type that is derived from gocty.
	// This is fine as we don't care about dynamic/unknown values, which don't exist in tf state.
	type TupleType struct {
		Int  int
		Map  map[string]string
		List []int
	}
	type NestObjectType struct {
		Field string `cty:"field"`
	}
	type ObjectType struct {
		Field int            `cty:"field"`
		Nest  NestObjectType `cty:"nest"`
	}
	type ValueGoType struct {
		AttrStr    string            `cty:"attr_str"`
		AttrInt    int               `cty:"attr_int"`
		AttrUint   uint              `cty:"attr_uint"`
		AttrFloat  float64           `cty:"attr_float"`
		AttrNumber big.Float         `cty:"attr_number"`
		AttrBool   bool              `cty:"attr_bool"`
		AttrList   []int             `cty:"attr_list"`
		AttrSet    []int             `cty:"attr_set"`
		AttrMap    map[string]string `cty:"attr_map"`
		AttrTuple  TupleType         `cty:"attr_tuple"`
		AttrObject ObjectType        `cty:"object"`
	}

	expectResourceValue := ValueGoType{
		AttrStr:    "some string",
		AttrInt:    -1,
		AttrUint:   1,
		AttrFloat:  0.1,
		AttrNumber: *big.NewFloat(0.5),
		AttrBool:   true,
		AttrList:   []int{1, 2, 3},
		AttrSet:    []int{1, 2, 3},
		AttrMap:    map[string]string{"key": "value"},
		AttrTuple: TupleType{
			Int:  1,
			Map:  map[string]string{"foo": "bar"},
			List: []int{1, 2, 3},
		},
		AttrObject: ObjectType{
			Field: 1,
			Nest: NestObjectType{
				Field: "a",
			},
		},
	}

	actual, err := tfstate.FromJSONStateResource(state, schemas)
	require.NoError(t, err)

	var actualResourceValue ValueGoType
	require.NoError(t, gocty.FromCtyValue(actual.Value, &actualResourceValue))

	actual.Value = cty.NilVal
	require.Equal(t, expectResourceWithoutValue, actual)

	av, ev := expectResourceValue, actualResourceValue
	require.Equal(t, ev.AttrStr, av.AttrStr)
	require.Equal(t, ev.AttrBool, av.AttrBool)
	require.Equal(t, ev.AttrInt, av.AttrInt)
	require.Equal(t, ev.AttrUint, av.AttrUint)
	require.Equal(t, ev.AttrFloat, av.AttrFloat)
	{
		var diff big.Float
		diff.Sub(&ev.AttrNumber, &av.AttrNumber)
		var abs big.Float
		abs.Abs(&diff)
		diffVal, _ := abs.Float64()
		require.Less(t, diffVal, 0.00001)
	}
	require.Equal(t, ev.AttrList, av.AttrList)
	require.Equal(t, ev.AttrSet, av.AttrSet)
	require.Equal(t, ev.AttrMap, av.AttrMap)
	require.Equal(t, ev.AttrTuple, av.AttrTuple)
	require.Equal(t, ev.AttrObject, av.AttrObject)
}

func TestFromJSONState(t *testing.T) {
	cases := []struct {
		name    string
		state   *tfjson.State
		schemas *tfjson.ProviderSchemas
		expect  *tfstate.State
		err     error
	}{
		{
			name:   "No values",
			state:  &tfjson.State{},
			expect: &tfstate.State{},
		},
		{
			name: "Empty values",
			state: &tfjson.State{
				Values: &tfjson.StateValues{},
			},
			expect: &tfstate.State{
				Values: &tfstate.StateValues{},
			},
		},
		{
			name: "Empty root module & output",
			state: &tfjson.State{
				Values: &tfjson.StateValues{
					RootModule: &tfjson.StateModule{},
					Outputs:    map[string]*tfjson.StateOutput{},
				},
			},
			expect: &tfstate.State{
				Values: &tfstate.StateValues{
					RootModule: &tfstate.StateModule{},
					Outputs:    map[string]*tfstate.StateOutput{},
				},
			},
		},
		{
			name: "One resource with outputs",
			state: &tfjson.State{
				Values: &tfjson.StateValues{
					RootModule: &tfjson.StateModule{
						Address: "root",
						Resources: []*tfjson.StateResource{
							{
								Address:      "demo_resource_foo.test",
								Mode:         tfjson.ManagedResourceMode,
								Type:         "demo_resource_foo",
								Name:         "test",
								Index:        1,
								ProviderName: "registry.terraform.io/magodo/demo",
								AttributeValues: map[string]interface{}{
									"attr_str": "some string",
								},
								SensitiveValues: json.RawMessage{
									1,
									2,
									3,
								},
								DependsOn: []string{
									"dep",
								},
								Tainted:    true,
								DeposedKey: "key",
							},
						},
						ChildModules: []*tfjson.StateModule{},
					},
					Outputs: map[string]*tfjson.StateOutput{
						"out": {
							Sensitive: true,
							Value:     1,
						},
					},
				},
			},
			schemas: &tfjson.ProviderSchemas{
				Schemas: map[string]*tfjson.ProviderSchema{
					"registry.terraform.io/magodo/demo": {
						ResourceSchemas: map[string]*tfjson.Schema{
							"demo_resource_foo": {
								Block: &tfjson.SchemaBlock{
									Attributes: map[string]*tfjson.SchemaAttribute{
										"attr_str": {
											AttributeType: cty.String,
										},
									},
								},
							},
						},
					},
				},
			},
			expect: &tfstate.State{
				Values: &tfstate.StateValues{
					RootModule: &tfstate.StateModule{
						Address: "root",
						Resources: []*tfstate.StateResource{
							{
								Address:      "demo_resource_foo.test",
								Mode:         tfjson.ManagedResourceMode,
								Type:         "demo_resource_foo",
								Name:         "test",
								Index:        1,
								ProviderName: "registry.terraform.io/magodo/demo",
								Value: cty.ObjectVal(map[string]cty.Value{
									"attr_str": cty.StringVal("some string"),
								}),
								SensitiveValues: json.RawMessage{
									1,
									2,
									3,
								},
								DependsOn: []string{
									"dep",
								},
								Tainted:    true,
								DeposedKey: "key",
							},
						},
					},
					Outputs: map[string]*tfstate.StateOutput{
						"out": {
							Sensitive: true,
							Value:     1,
						},
					},
				},
			},
			err: nil,
		},
		{
			name: "Nested module",
			state: &tfjson.State{
				Values: &tfjson.StateValues{
					RootModule: &tfjson.StateModule{
						Address: "root",
						ChildModules: []*tfjson.StateModule{
							{
								Address: "child",
								Resources: []*tfjson.StateResource{
									{
										Address:      "demo_resource_foo.test",
										Mode:         tfjson.ManagedResourceMode,
										Type:         "demo_resource_foo",
										Name:         "test",
										Index:        1,
										ProviderName: "registry.terraform.io/magodo/demo",
										AttributeValues: map[string]interface{}{
											"attr_str": "some string",
										},
										SensitiveValues: json.RawMessage{
											1,
											2,
											3,
										},
										DependsOn: []string{
											"dep",
										},
										Tainted:    true,
										DeposedKey: "key",
									},
								},
							},
						},
					},
					Outputs: map[string]*tfjson.StateOutput{
						"out": {
							Sensitive: true,
							Value:     1,
						},
					},
				},
			},
			schemas: &tfjson.ProviderSchemas{
				Schemas: map[string]*tfjson.ProviderSchema{
					"registry.terraform.io/magodo/demo": {
						ResourceSchemas: map[string]*tfjson.Schema{
							"demo_resource_foo": {
								Block: &tfjson.SchemaBlock{
									Attributes: map[string]*tfjson.SchemaAttribute{
										"attr_str": {
											AttributeType: cty.String,
										},
									},
								},
							},
						},
					},
				},
			},
			expect: &tfstate.State{
				Values: &tfstate.StateValues{
					RootModule: &tfstate.StateModule{
						Address: "root",
						ChildModules: []*tfstate.StateModule{
							{
								Address: "child",
								Resources: []*tfstate.StateResource{
									{
										Address:      "demo_resource_foo.test",
										Mode:         tfjson.ManagedResourceMode,
										Type:         "demo_resource_foo",
										Name:         "test",
										Index:        1,
										ProviderName: "registry.terraform.io/magodo/demo",
										Value: cty.ObjectVal(map[string]cty.Value{
											"attr_str": cty.StringVal("some string"),
										}),
										SensitiveValues: json.RawMessage{
											1,
											2,
											3,
										},
										DependsOn: []string{
											"dep",
										},
										Tainted:    true,
										DeposedKey: "key",
									},
								},
							},
						},
					},
					Outputs: map[string]*tfstate.StateOutput{
						"out": {
							Sensitive: true,
							Value:     1,
						},
					},
				},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := tfstate.FromJSONState(c.state, c.schemas)
			if c.err != nil {
				require.Errorf(t, err, c.err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.expect, actual)
		})
	}
}
