package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

type RealtimeProtectionPolicy struct{}

type RealtimeProtectionPolicyArgs struct {
	TenantURL       string              `pulumi:"tenantUrl"`
	APIToken        *string             `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken     *string             `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode        *string             `pulumi:"authMode,optional"`
	OAuth2          *NetskopeOAuth2Args `pulumi:"oauth2,optional"`
	Name            string              `pulumi:"name"`
	PolicyGroupID   *int                `pulumi:"policyGroupId,optional"`
	PolicyGroupName *string             `pulumi:"policyGroupName,optional"`
	AppIDs          []int               `pulumi:"appIds,optional"`
	AppTags         []string            `pulumi:"appTags,optional"`
	Users           []string            `pulumi:"users,optional"`
	Groups          []string            `pulumi:"groups,optional"`
	Action          string              `pulumi:"action"`
	Enabled         bool                `pulumi:"enabled"`
}

type RealtimeProtectionPolicyOutputs struct {
	RealtimeProtectionPolicyArgs
	PolicyID              int `pulumi:"policyId"`
	ResolvedPolicyGroupID int `pulumi:"resolvedPolicyGroupId"`
}

func (*RealtimeProtectionPolicy) Annotate(a infer.Annotator) {
	a.SetToken("index", "RealtimeProtectionPolicy")
}

func (*RealtimeProtectionPolicy) Create(ctx context.Context, req infer.CreateRequest[RealtimeProtectionPolicyArgs]) (infer.CreateResponse[RealtimeProtectionPolicyOutputs], error) {
	output, payload, err := realtimePolicyPayloadFromArgs(ctx, req.Inputs)
	if err != nil {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	if req.DryRun {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{ID: req.Inputs.Name, Output: output}, nil
	}

	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	created, err := client.createRealtimePolicy(ctx, payload)
	if err != nil {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	output.PolicyID = created.ID
	return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{ID: strconv.Itoa(created.ID), Output: output}, nil
}

func (*RealtimeProtectionPolicy) Read(ctx context.Context, req infer.ReadRequest[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]) (infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs], error) {
	return infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{ID: req.ID, Inputs: req.Inputs, State: req.State}, nil
}

func (*RealtimeProtectionPolicy) Update(ctx context.Context, req infer.UpdateRequest[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]) (infer.UpdateResponse[RealtimeProtectionPolicyOutputs], error) {
	output, payload, err := realtimePolicyPayloadFromArgs(ctx, req.Inputs)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	policyID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, fmt.Errorf("invalid realtime policy ID %q: %w", req.ID, err)
	}
	output.PolicyID = policyID
	if req.DryRun {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{Output: output}, nil
	}

	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	updated, err := client.updateRealtimePolicy(ctx, policyID, payload)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	output.PolicyID = updated.ID
	return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{Output: output}, nil
}

func (*RealtimeProtectionPolicy) Delete(ctx context.Context, req infer.DeleteRequest[RealtimeProtectionPolicyOutputs]) (infer.DeleteResponse, error) {
	policyID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid realtime policy ID %q: %w", req.ID, err)
	}
	client := newResourceClient(req.State.TenantURL, req.State.APIToken, req.State.BearerToken, req.State.AuthMode, req.State.OAuth2, http.DefaultClient)
	return infer.DeleteResponse{}, client.deleteRealtimePolicy(ctx, policyID)
}

func realtimePolicyPayloadFromArgs(ctx context.Context, args RealtimeProtectionPolicyArgs) (RealtimeProtectionPolicyOutputs, realtimePolicyPayload, error) {
	output := RealtimeProtectionPolicyOutputs{RealtimeProtectionPolicyArgs: args}
	groupID := 0
	if args.PolicyGroupID != nil {
		groupID = *args.PolicyGroupID
	}
	if groupID == 0 && args.PolicyGroupName != nil && *args.PolicyGroupName != "" {
		client := newResourceClient(args.TenantURL, args.APIToken, args.BearerToken, args.AuthMode, args.OAuth2, http.DefaultClient)
		group, err := client.findPolicyGroupByName(ctx, *args.PolicyGroupName)
		if err != nil {
			return output, realtimePolicyPayload{}, err
		}
		if group == nil {
			return output, realtimePolicyPayload{}, fmt.Errorf("policy group %q not found", *args.PolicyGroupName)
		}
		groupID = group.ID
	}

	output.ResolvedPolicyGroupID = groupID
	return output, realtimePolicyPayload{
		Name:          args.Name,
		PolicyGroupID: groupID,
		AppIDs:        args.AppIDs,
		AppTags:       args.AppTags,
		Users:         args.Users,
		Groups:        args.Groups,
		Action:        args.Action,
		Enabled:       args.Enabled,
	}, nil
}
