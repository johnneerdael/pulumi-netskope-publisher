package provider

import (
	"encoding/base64"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func plainUserData(payload pulumi.StringOutput) pulumi.StringOutput {
	return payload
}

func base64UserData(payload pulumi.StringOutput) pulumi.StringOutput {
	return payload.ApplyT(func(value string) string {
		return base64.StdEncoding.EncodeToString([]byte(value))
	}).(pulumi.StringOutput)
}

func metadataUserData(payload pulumi.StringOutput, key string) pulumi.Map {
	return pulumi.Map{key: payload}
}

func base64MetadataUserData(payload pulumi.StringOutput, key string) pulumi.Map {
	return pulumi.Map{key: base64UserData(payload)}
}

func guestInfoUserData(payload pulumi.StringOutput) pulumi.Map {
	return pulumi.Map{
		"guestinfo.userdata":          base64UserData(payload),
		"guestinfo.userdata.encoding": pulumi.String("base64"),
	}
}

func scalewayUserData(payload pulumi.StringOutput) pulumi.Map {
	return pulumi.Map{
		"cloudInit": payload,
		"userData":  pulumi.Map{"cloud-init": payload},
	}
}
