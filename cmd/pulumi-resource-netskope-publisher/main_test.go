package main

import (
	"encoding/json"
	"os/exec"
	"testing"
)

func TestSchemaCommandReturnsProviderSchema(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--schema")
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("schema command failed: %v\n%s", err, output)
	}

	var schema struct {
		Name      string                 `json:"name"`
		Provider  map[string]any         `json:"provider"`
		Resources map[string]interface{} `json:"resources"`
	}
	if err := json.Unmarshal(output, &schema); err != nil {
		t.Fatalf("schema command did not return JSON: %v\n%s", err, output)
	}

	if schema.Name != "netskope-publisher" {
		t.Fatalf("schema name = %q, want netskope-publisher", schema.Name)
	}

	for _, token := range []string{
		"netskope-publisher:index:AwsPublisher",
		"netskope-publisher:index:AzurePublisher",
		"netskope-publisher:index:GcpPublisher",
		"netskope-publisher:index:VspherePublisher",
		"netskope-publisher:index:HypervPublisher",
	} {
		if _, ok := schema.Resources[token]; !ok {
			t.Fatalf("schema missing component token %s", token)
		}
	}
}
