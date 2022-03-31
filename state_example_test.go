package tfstate_test

import (
	"context"
	"log"

	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/magodo/tfstate"
)

func ExampleFromJSONState() {
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
	schema, err := tf.ProvidersSchema(ctx)
	if err != nil {
		log.Fatalf("get provider schema: %v", err)
	}
	rawState, err := tf.Show(ctx)
	if err != nil {
		log.Fatalf("show state: %v", err)
	}
	state, err := tfstate.FromJSONState(rawState, schema)
	if err != nil {
		log.Fatal(err)
	}
	_ = state
}
