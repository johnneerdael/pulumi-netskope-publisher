# NPA Application Resources Design

## Goal

Broaden the provider from publisher deployment only into a focused Netskope
Private Access deployment provider. The expanded provider should let Pulumi
create the access path for an application: deploy publishers, register private
applications, tag those applications, create the NPA realtime protection rule,
and reconcile which Pulumi-managed publishers serve applications with matching
private app tags.

The first implementation should stay deployment-focused. It should not attempt
to model every Netskope administrative endpoint exposed by the MCP project or
Swagger file.

## Current Context

The active provider is implemented in Go with `pulumi-go-provider`. It exposes a
stateful `NetskopeRegistration` resource and many publisher components. The
generated SDKs come from the provider schema.

The nearby `../privateaccess-mcp` project already proves the needed Netskope API
surface:

- Private app CRUD: `/api/v2/steering/apps/private`
- Private app tags: `/api/v2/steering/apps/private/tags`
- Private app publisher associations:
  `/api/v2/steering/apps/private/publishers`
- NPA policy rules: `/api/v2/policy/npa/rules`
- Policy groups: `/api/v2/policy/npa/policygroups`

Use that project and `/Users/jneerdael/Scripts/privateaccess-mcp/swagger.json`
as reference material for request and response shapes. Do not copy the MCP
runtime architecture into the Pulumi provider.

## Chosen Scope

Add a focused NPA core resource layer:

- `PrivateApp`
- Inline private app tags on `PrivateApp`
- Optional separate private app tag attachment resource if shared ownership
  needs it
- `RealtimeProtectionPolicy` for NPA realtime protection rules
- `TagPublisherAssignment` for app-tag-to-publisher-pool reconciliation
- Publisher component `placementLabels`

Leave these out of the first pass unless a deployment workflow later requires
them:

- Alerts
- Discovery settings
- Publisher upgrade profiles
- Broad policy group lifecycle beyond lookup/reference
- SCIM
- Reporting
- General steering administration

## Resource Ownership

Use normal Pulumi ownership semantics for resources Pulumi creates.

`PrivateApp` and `RealtimeProtectionPolicy` are authoritative for the fields
they own. Updates reconcile remote state to the Pulumi inputs. Deletes remove
the owned remote object.

Shared association resources are scoped and additive outside their scope.
`TagPublisherAssignment` may add and remove only the publisher IDs selected by
that resource. It must leave unrelated publisher assignments, unrelated tags,
private app objects, and publisher objects untouched.

## PrivateApp Resource

`PrivateApp` is a stateful custom resource backed by
`/api/v2/steering/apps/private`.

Inputs should cover the fields needed for deployment-time app registration:

- `tenantUrl`
- `apiToken`, `bearerToken`, `authMode`, and `oauth2`
- `appName`
- `appType`: `client` or `clientless`
- `host` or `hosts`
- `protocols`
- `clientlessAccess`
- `isUserPortalApp`
- `usePublisherDns`
- `trustSelfSignedCerts`
- `tags` or `privateAppTags`
- `adoptExisting`

Create-only is the default. If a same-name app already exists, create should
fail with a clear message explaining that the user can import the resource or
set `adoptExisting: true`.

When `adoptExisting: true`, the resource may find an existing app by name,
record its ID, and then manage the declared fields authoritatively.

Dry runs should return stable planned outputs without mutating Netskope.

## Private App Tags

The common path should be inline tags on `PrivateApp`.

If a separate attachment resource is added, it should be additive by default:
it adds or removes only its declared tags for its declared private apps. It
should not replace all tags unless an explicit authoritative mode is designed
and tested.

## RealtimeProtectionPolicy Resource

`RealtimeProtectionPolicy` is a stateful custom resource for one NPA policy
rule backed by `/api/v2/policy/npa/rules`.

It should support:

- Existing policy group lookup or reference by ID/name
- Private app IDs and/or private app tag references
- Users and groups where the API supports them
- Action
- Enabled state
- Ordering or priority only where the API exposes it cleanly

The first pass should not require full policy group lifecycle management.
Policy group creation can be added later if it becomes part of the deployment
workflow.

## Publisher Placement Labels

Netskope publishers do not support the same tags as private apps. Matching
private apps to publisher pools therefore needs Pulumi-side metadata.

Each publisher component should accept user-defined placement labels:

```ts
placementLabels: ["vpc-a", "production-eu"]
```

Use list labels rather than key/value labels in the first design. They are
portable across all supported providers and do not require normalizing cloud
network identifiers across AWS, Azure, GCP, vSphere, Kubernetes, and other
platforms.

Publisher components should expose these labels in outputs alongside the
publisher IDs from registration data. The labels are Pulumi metadata only; they
are not written to Netskope publisher objects.

## TagPublisherAssignment Resource

`TagPublisherAssignment` reconciles private app tags to a Pulumi-managed
publisher pool.

Example shape:

```ts
new TagPublisherAssignment("vpc-a-access", {
  tenantUrl,
  bearerToken,
  appTags: ["vpc-a"],
  publisherPlacementLabels: ["vpc-a"],
  publishers: publisher.publishers,
});
```

On create or update:

1. List private apps from Netskope.
2. Filter apps by the configured `appTags`.
3. Select Pulumi-managed publisher IDs whose `placementLabels` match
   `publisherPlacementLabels`.
4. Add selected publisher IDs to matching apps.
5. Remove selected publisher IDs from apps that no longer match.
6. Leave publisher IDs outside this resource's selected set untouched.
7. Leave private app tags untouched.

Default tag matching should be `any`. A later `matchMode` input can support
`all` if needed.

On delete, the resource may remove only the selected publisher IDs from apps it
last managed if state can track that safely. It must not delete private apps,
tags, or publisher objects.

Publisher object lifecycle stays with the publisher components. A publisher
with no associated private apps is not removed by `TagPublisherAssignment`.

Adding a new Pulumi-managed publisher with placement label `vpc-a` should update
the publisher output set. The next Pulumi update should reconcile all private
apps tagged `vpc-a` to include that publisher.

## API Client

Refactor the current registration-specific Netskope client into a shared Go
client that supports:

- Publisher registration endpoints already used by `NetskopeRegistration`
- Private app CRUD and list
- Private app tag endpoints
- Private app publisher association endpoints
- NPA policy rule endpoints
- Policy group lookup endpoints
- Existing token and OAuth2 authentication behavior

Keep the client small and endpoint-specific. Avoid generating a full Swagger
client for this pass.

## Implementation Files

Keep the provider implementation in `internal/provider/`.

Likely files:

- `netskope_client.go`
- `private_app.go`
- `private_app_tags.go`, only if a separate attachment resource is included
- `tag_publisher_assignment.go`
- `realtime_protection_policy.go`

Register the new resources in `internal/provider/provider.go`. Continue using
the Go provider structs as the schema source for generated SDKs.

## Testing

Use `httptest` like the current `NetskopeRegistration` tests.

Required tests:

- `PrivateApp` creates, updates, deletes, and adopts existing apps.
- Creating a `PrivateApp` fails on a same-name existing app unless
  `adoptExisting` is true.
- Private app tag handling preserves unrelated tags when using additive
  attachment semantics.
- `TagPublisherAssignment` adds selected publisher IDs to matching apps.
- `TagPublisherAssignment` removes only selected publisher IDs from apps that
  no longer match.
- `TagPublisherAssignment` leaves unrelated publisher IDs untouched.
- `RealtimeProtectionPolicy` creates, updates, and deletes NPA policy rules.
- OAuth2 and token authentication behavior is shared across old and new
  resources.
- Dry-run paths return stable outputs without API mutation.

## Documentation

Update docs and examples around the new deployment story:

- Publisher components can be labeled with `placementLabels`.
- Private applications can be registered from the same Pulumi program that
  deploys app infrastructure.
- Application tags select which private apps belong to a logical network or
  placement pool.
- `TagPublisherAssignment` reconciles app tags to Pulumi-managed publishers.
- Realtime protection policies make the registered private app usable for the
  intended users or groups.

The provider name and package identity may need a later naming decision. The
implementation can begin under the existing package without making a breaking
rename.
