package state

import (
	"encoding/json"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/magodo/tfadd/tfadd/terraform/jsonschema"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type ResourceMode string

const (
	DataResourceMode    ResourceMode = "data"
	ManagedResourceMode ResourceMode = "managed"
)

type State struct {
	TerraformVersion string
	Values           *StateValues
}

type StateValues struct {
	RootModule *StateModule
	Outputs    map[string]*StateOutput
}

type StateOutput struct {
	Sensitive bool
	Value     interface{}
}

type StateModule struct {
	Resources    []*StateResource
	Address      string
	ChildModules []*StateModule
}

type StateResource struct {
	Address       string
	Mode          ResourceMode
	Type          string
	Name          string
	Index         interface{}
	ProviderName  string
	SchemaVersion uint64
	Value         cty.Value
	DependsOn     []string
	Tainted       bool
	DeposedKey    string
}

func FromJSONState(rawState *tfjson.State, providerSchema *tfjson.ProviderSchema) (*State, error) {
	if rawState == nil {
		return nil, nil
	}
	if providerSchema == nil {
		return nil, fmt.Errorf("provider schema cannot be nil")
	}
	state := &State{
		TerraformVersion: rawState.FormatVersion,
	}
	if rawState.Values == nil {
		return state, nil
	}
	rootModule, err := fromJSONStateModule(rawState.Values.RootModule, providerSchema)
	if err != nil {
		return nil, err
	}
	state.Values = &StateValues{
		RootModule: rootModule,
	}
	if size := len(rawState.Values.Outputs); size > 0 {
		m := make(map[string]*StateOutput, size)
		for name, output := range rawState.Values.Outputs {
			m[name] = fromJSONStateOutput(output, providerSchema)
		}
		state.Values.Outputs = m
	}
	return state, nil
}

func fromJSONStateModule(module *tfjson.StateModule, providerSchema *tfjson.ProviderSchema) (*StateModule, error) {
	if module == nil {
		return nil, nil
	}
	ret := &StateModule{
		Address: module.Address,
	}
	var err error
	if size := len(module.Resources); size > 0 {
		resources := make([]*StateResource, size)
		for i, resource := range module.Resources {
			resources[i], err = fromJSONStateResource(resource, providerSchema)
			if err != nil {
				return nil, fmt.Errorf("converting json state for resource: %v", err)
			}
		}
		ret.Resources = resources
	}
	if size := len(module.ChildModules); size > 0 {
		modules := make([]*StateModule, size)
		for i, module := range module.ChildModules {
			modules[i], err = fromJSONStateModule(module, providerSchema)
			if err != nil {
				return nil, fmt.Errorf("converting json state for module: %v", err)
			}
		}
		ret.ChildModules = modules
	}
	return ret, nil
}

func fromJSONStateOutput(output *tfjson.StateOutput, providerSchema *tfjson.ProviderSchema) *StateOutput {
	if output == nil {
		return nil
	}
	return &StateOutput{
		Sensitive: output.Sensitive,
		Value:     output.Value,
	}
}

func fromJSONStateResource(resource *tfjson.StateResource, providerSchema *tfjson.ProviderSchema) (*StateResource, error) {
	resourceSchema, ok := providerSchema.ResourceSchemas[resource.Type]
	if !ok {
		return nil, fmt.Errorf("No resource type %q found in the provider schema", resource.Type)
	}
	if resource == nil {
		return nil, nil
	}
	ret := &StateResource{
		Address:       resource.Address,
		Mode:          ResourceMode(resource.Mode),
		Type:          resource.Type,
		Name:          resource.Name,
		Index:         resource.Index,
		ProviderName:  resource.ProviderName,
		SchemaVersion: resource.SchemaVersion,
		DependsOn:     resource.DependsOn,
		Tainted:       resource.Tainted,
		DeposedKey:    resource.DeposedKey,
	}
	attrsJSON, err := json.Marshal(resource.AttributeValues)
	if err != nil {
		return nil, fmt.Errorf("marshal %q: %v", resource.AttributeValues, err)
	}
	val, err := ctyjson.Unmarshal(attrsJSON, jsonschema.SchemaBlockImpliedType(resourceSchema.Block))
	if err != nil {
		return nil, fmt.Errorf("cty json unmarshal %q: %v", attrsJSON, err)
	}
	ret.Value = val
	return ret, nil
}
