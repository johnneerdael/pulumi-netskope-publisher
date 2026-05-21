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
	output := RealtimeProtectionPolicyOutputs{RealtimeProtectionPolicyArgs: req.Inputs}
	if req.DryRun {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{ID: req.Inputs.Name, Output: output}, nil
	}

	output, payload, err := realtimePolicyPayloadFromArgs(ctx, req.Inputs)
	if err != nil {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}

	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	created, err := client.createRealtimePolicy(ctx, payload)
	if err != nil {
		return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	output.PolicyID = created.RuleID
	return infer.CreateResponse[RealtimeProtectionPolicyOutputs]{ID: strconv.Itoa(created.RuleID), Output: output}, nil
}

func (*RealtimeProtectionPolicy) Read(ctx context.Context, req infer.ReadRequest[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]) (infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs], error) {
	policyID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{}, fmt.Errorf("invalid realtime policy ID %q: %w", req.ID, err)
	}
	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	policy, err := client.getRealtimePolicy(ctx, policyID)
	if err != nil {
		if err == errNetskopeNotFound {
			return infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{}, nil
		}
		return infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{}, err
	}
	state := req.State
	state.PolicyID = policy.RuleID
	return infer.ReadResponse[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]{ID: strconv.Itoa(state.PolicyID), Inputs: req.Inputs, State: state}, nil
}

func (*RealtimeProtectionPolicy) Update(ctx context.Context, req infer.UpdateRequest[RealtimeProtectionPolicyArgs, RealtimeProtectionPolicyOutputs]) (infer.UpdateResponse[RealtimeProtectionPolicyOutputs], error) {
	output := RealtimeProtectionPolicyOutputs{RealtimeProtectionPolicyArgs: req.Inputs}
	policyID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, fmt.Errorf("invalid realtime policy ID %q: %w", req.ID, err)
	}
	output.PolicyID = policyID
	if req.DryRun {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{Output: output}, nil
	}

	output, payload, err := realtimePolicyPayloadFromArgs(ctx, req.Inputs)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	output.PolicyID = policyID

	client := newResourceClient(req.Inputs.TenantURL, req.Inputs.APIToken, req.Inputs.BearerToken, req.Inputs.AuthMode, req.Inputs.OAuth2, http.DefaultClient)
	updated, err := client.updateRealtimePolicy(ctx, policyID, payload)
	if err != nil {
		return infer.UpdateResponse[RealtimeProtectionPolicyOutputs]{}, err
	}
	output.PolicyID = updated.RuleID
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

	groupIDString := ""
	if groupID != 0 {
		groupIDString = strconv.Itoa(groupID)
	}

	output.ResolvedPolicyGroupID = groupID
	return output, realtimePolicyPayload{
		RuleName: args.Name,
		GroupID:  groupIDString,
		RuleData: realtimePolicyRuleData{
			PrivateApps:    intStrings(args.AppIDs),
			PrivateAppTags: args.AppTags,
			Users:          args.Users,
			UserGroups:     args.Groups,
			MatchCriteriaAction: realtimePolicyAction{
				ActionName: args.Action,
			},
		},
		Enabled: enabledString(args.Enabled),
	}, nil
}

func intStrings(values []int) []string {
	strings := make([]string, 0, len(values))
	for _, value := range values {
		strings = append(strings, strconv.Itoa(value))
	}
	return strings
}

func enabledString(enabled bool) string {
	if enabled {
		return "1"
	}
	return "0"
}
