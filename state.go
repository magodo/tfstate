package tfstate

import (
	"encoding/json"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/magodo/tfstate/internal/terraform/jsonschema"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

const stateVersionSupported = "4"

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
	Address         string
	Mode            tfjson.ResourceMode
	Type            string
	Name            string
	Index           interface{}
	ProviderName    string
	SchemaVersion   uint64
	Value           cty.Value
	SensitiveValues json.RawMessage
	DependsOn       []string
	Tainted         bool
	DeposedKey      string
}

func FromJSONState(rawState *tfjson.State, schemas *tfjson.ProviderSchemas) (*State, error) {
	if rawState == nil {
		return nil, nil
	}
	if rawState.FormatVersion != stateVersionSupported {
		return nil, fmt.Errorf("tfstate only supports state version %s. got=%s", stateVersionSupported, rawState.FormatVersion)
	}
	state := &State{
		TerraformVersion: rawState.FormatVersion,
	}
	if rawState.Values == nil {
		return state, nil
	}
	state.Values = &StateValues{}
	if rawState.Values.RootModule != nil {
		rootModule, err := FromJSONStateModule(rawState.Values.RootModule, schemas)
		if err != nil {
			return nil, err
		}
		state.Values = &StateValues{
			RootModule: rootModule,
		}
	}
	if rawState.Values.Outputs != nil {
		m := make(map[string]*StateOutput, len(rawState.Values.Outputs))
		for name, output := range rawState.Values.Outputs {
			m[name] = FromJSONStateOutput(output)
		}
		state.Values.Outputs = m
	}
	return state, nil
}

func FromJSONStateModule(module *tfjson.StateModule, schemas *tfjson.ProviderSchemas) (*StateModule, error) {
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
			resources[i], err = FromJSONStateResource(resource, schemas)
			if err != nil {
				return nil, fmt.Errorf("converting json state for resource: %v", err)
			}
		}
		ret.Resources = resources
	}
	if size := len(module.ChildModules); size > 0 {
		modules := make([]*StateModule, size)
		for i, module := range module.ChildModules {
			modules[i], err = FromJSONStateModule(module, schemas)
			if err != nil {
				return nil, fmt.Errorf("converting json state for module: %v", err)
			}
		}
		ret.ChildModules = modules
	}
	return ret, nil
}

func FromJSONStateOutput(output *tfjson.StateOutput) *StateOutput {
	if output == nil {
		return nil
	}
	return &StateOutput{
		Sensitive: output.Sensitive,
		Value:     output.Value,
	}
}

func FromJSONStateResource(resource *tfjson.StateResource, schemas *tfjson.ProviderSchemas) (*StateResource, error) {
	if schemas == nil {
		return nil, fmt.Errorf("provider schemas is nil")
	}
	if schemas.Schemas == nil {
		return nil, fmt.Errorf("provider schemas' Schemas is nil")
	}
	providerSchema, ok := schemas.Schemas[resource.ProviderName]
	if !ok {
		return nil, fmt.Errorf("No provider type %q found in the provider schemas", resource.ProviderName)
	}
	resourceSchema, ok := providerSchema.ResourceSchemas[resource.Type]
	if !ok {
		return nil, fmt.Errorf("No resource type %q found in the provider schema", resource.Type)
	}
	if resource == nil {
		return nil, nil
	}
	ret := &StateResource{
		Address:         resource.Address,
		Mode:            resource.Mode,
		Type:            resource.Type,
		Name:            resource.Name,
		Index:           resource.Index,
		ProviderName:    resource.ProviderName,
		SchemaVersion:   resource.SchemaVersion,
		SensitiveValues: resource.SensitiveValues,
		DependsOn:       resource.DependsOn,
		Tainted:         resource.Tainted,
		DeposedKey:      resource.DeposedKey,
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
