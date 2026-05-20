import assert from "node:assert/strict";
import test from "node:test";
import * as pulumi from "@pulumi/pulumi";
import { RawResource } from "../src/rawResource";

test("RawResource registers the requested token and inputs", async () => {
  const seen: Array<{ type: string; inputs: pulumi.Inputs }> = [];
  pulumi.runtime.setMocks({
    newResource(args) {
      seen.push({ type: args.type, inputs: args.inputs });
      return { id: `${args.name}-id`, state: args.inputs };
    },
    call(args) {
      return args.inputs;
    },
  });

  const resource = new RawResource("example", "example:index/server:Server", { userData: "#cloud-config" });
  await new Promise<string>((resolve) => resource.id.apply((id) => {
    resolve(id);
    return id;
  }));

  assert.equal(seen[0].type, "example:index/server:Server");
  assert.equal(seen[0].inputs.userData, "#cloud-config");
});
