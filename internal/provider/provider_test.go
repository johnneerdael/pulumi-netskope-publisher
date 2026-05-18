package provider

import (
	"strings"
	"testing"

	"github.com/blang/semver"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func TestConstructReturnsExplicitParityError(t *testing.T) {
	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	server, err := integration.NewServer(
		t.Context(),
		Name,
		semver.MustParse("0.1.0"),
		integration.WithProvider(provider),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = server.Construct(p.ConstructRequest{
		Urn: "urn:pulumi:stack::project::netskope-publisher:index:AwsPublisher::publisher",
		Inputs: property.NewMap(map[string]property.Value{
			"subnetId":         property.New("subnet-123"),
			"securityGroupIds": property.New([]property.Value{property.New("sg-123")}),
			"registrations":    property.New(map[string]property.Value{}),
			"namePrefix":       property.New("pub"),
			"replicas":         property.New(1.0),
			"tenantUrl":        property.New("https://example.goskope.com"),
			"apiToken":         property.New("token"),
		}),
	})
	if err == nil {
		t.Fatal("expected construct to return parity error")
	}
	if !strings.Contains(err.Error(), "Go provider child-resource parity is not implemented") {
		t.Fatalf("unexpected error: %v", err)
	}
}
