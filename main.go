package main

import (
	"context"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/magodo/tfadd/tfadd/state"
)

func main() {
	ctx := context.TODO()
	av := fs.AnyVersion{
		Product: &product.Terraform,
	}
	execPath, err := av.Find(ctx)
	if err != nil {
		log.Fatal(err)
	}
	tf, err := tfexec.NewTerraform(".", execPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := add(ctx, tf); err != nil {
		log.Fatal(err)
	}
}

func add(ctx context.Context, tf *tfexec.Terraform) error {
	schema, err := tf.ProvidersSchema(ctx)
	if err != nil {
		return fmt.Errorf("get provider schema: %v", err)
	}
	providerSchema := schema.Schemas["registry.terraform.io/hashicorp/azurerm"]
	rawState, err := tf.Show(ctx)
	if err != nil {
		return fmt.Errorf("show state: %v", err)
	}
	state, err := state.FromJSONState(rawState, providerSchema)
	if err != nil {
		return err
	}
	spew.Dump(state)
	return nil
}
