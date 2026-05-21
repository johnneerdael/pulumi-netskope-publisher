package provider

import (
	"reflect"
	"testing"
)

type catalogValidationTaggedArgs struct {
	TemplateID string `pulumi:"templateId"`
	VpcUUID    string `pulumi:"vpcUuid,optional"`
	OSID       *int   `pulumi:"osId,optional"`
}

func TestIsGoFieldMissingUsesPulumiTags(t *testing.T) {
	value := reflectValue(catalogValidationTaggedArgs{
		TemplateID: "template-1",
		VpcUUID:    "vpc-1",
	})

	if isGoFieldMissing(value, "templateId") {
		t.Fatal("expected templateId to resolve through the pulumi tag")
	}
	if isGoFieldMissing(value, "vpcUuid") {
		t.Fatal("expected vpcUuid to resolve through the pulumi tag")
	}
	if !isGoFieldMissing(value, "osId") {
		t.Fatal("expected nil pointer field osId to be missing")
	}
}

func TestIsGoFieldMissingTreatsUnknownPulumiTagAsMissing(t *testing.T) {
	value := reflectValue(catalogValidationTaggedArgs{
		TemplateID: "template-1",
	})

	if !isGoFieldMissing(value, "unknownField") {
		t.Fatal("expected unknown field to be missing")
	}
}

func reflectValue(value any) reflect.Value {
	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Pointer {
		reflected = reflected.Elem()
	}
	return reflected
}
