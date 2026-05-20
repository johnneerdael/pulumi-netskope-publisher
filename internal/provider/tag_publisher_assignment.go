package provider

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"
)

type TagPublisherAssignment struct{}

type PublisherAssignmentInput struct {
	PublisherID     int      `pulumi:"publisherId"`
	PlacementLabels []string `pulumi:"placementLabels,optional"`
}

type TagPublisherAssignmentArgs struct {
	TenantURL                string                              `pulumi:"tenantUrl"`
	APIToken                 *string                             `pulumi:"apiToken,optional" provider:"secret"`
	BearerToken              *string                             `pulumi:"bearerToken,optional" provider:"secret"`
	AuthMode                 *string                             `pulumi:"authMode,optional"`
	OAuth2                   *NetskopeOAuth2Args                 `pulumi:"oauth2,optional"`
	AppTags                  []string                            `pulumi:"appTags"`
	PublisherPlacementLabels []string                            `pulumi:"publisherPlacementLabels"`
	Publishers               map[string]PublisherAssignmentInput `pulumi:"publishers"`
	MatchMode                *string                             `pulumi:"matchMode,optional"`
}

type TagPublisherAssignmentOutputs struct {
	TagPublisherAssignmentArgs
	MatchedApps        []string `pulumi:"matchedApps"`
	SelectedPublishers []int    `pulumi:"selectedPublishers"`
}

func (*TagPublisherAssignment) Annotate(a infer.Annotator) {
	a.SetToken("index", "TagPublisherAssignment")
}

func (*TagPublisherAssignment) Create(ctx context.Context, req infer.CreateRequest[TagPublisherAssignmentArgs]) (infer.CreateResponse[TagPublisherAssignmentOutputs], error) {
	output, err := reconcileTagPublisherAssignment(ctx, req.Inputs, req.DryRun)
	if err != nil {
		return infer.CreateResponse[TagPublisherAssignmentOutputs]{}, err
	}
	return infer.CreateResponse[TagPublisherAssignmentOutputs]{ID: strings.Join(req.Inputs.AppTags, ","), Output: output}, nil
}

func (*TagPublisherAssignment) Read(ctx context.Context, req infer.ReadRequest[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs]) (infer.ReadResponse[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs], error) {
	return infer.ReadResponse[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs]{ID: req.ID, Inputs: req.Inputs, State: req.State}, nil
}

func (*TagPublisherAssignment) Update(ctx context.Context, req infer.UpdateRequest[TagPublisherAssignmentArgs, TagPublisherAssignmentOutputs]) (infer.UpdateResponse[TagPublisherAssignmentOutputs], error) {
	output, err := reconcileTagPublisherAssignment(ctx, req.Inputs, req.DryRun)
	if err != nil {
		return infer.UpdateResponse[TagPublisherAssignmentOutputs]{}, err
	}
	return infer.UpdateResponse[TagPublisherAssignmentOutputs]{Output: output}, nil
}

func (*TagPublisherAssignment) Delete(ctx context.Context, req infer.DeleteRequest[TagPublisherAssignmentOutputs]) (infer.DeleteResponse, error) {
	return infer.DeleteResponse{}, nil
}

func reconcileTagPublisherAssignment(ctx context.Context, args TagPublisherAssignmentArgs, dryRun bool) (TagPublisherAssignmentOutputs, error) {
	selected := selectPublishersByPlacement(args.Publishers, args.PublisherPlacementLabels)
	output := TagPublisherAssignmentOutputs{
		TagPublisherAssignmentArgs: args,
		SelectedPublishers:         selected,
	}
	if dryRun {
		return output, nil
	}

	client := newResourceClient(args.TenantURL, args.APIToken, args.BearerToken, args.AuthMode, args.OAuth2, http.DefaultClient)
	apps, err := client.listPrivateAppsWithPublishers(ctx)
	if err != nil {
		return output, err
	}

	selectedSet := intSet(selected)
	for _, app := range apps {
		matches := appMatchesTags(app.Tags, args.AppTags, defaultString(args.MatchMode, "any"))
		current := currentPublisherIDs(app.ServicePublisherAssignments)
		next := reconcilePublisherIDs(current, selectedSet, matches)
		if !sameInts(current, next) {
			if err := client.replacePrivateAppPublishers(ctx, []string{app.AppName}, next); err != nil {
				return output, err
			}
		}
		if matches {
			output.MatchedApps = append(output.MatchedApps, app.AppName)
		}
	}
	sort.Strings(output.MatchedApps)
	return output, nil
}

func selectPublishersByPlacement(publishers map[string]PublisherAssignmentInput, labels []string) []int {
	labelSet := stringSet(labels)
	var selected []int
	for _, publisher := range publishers {
		if intersectsStringSet(publisher.PlacementLabels, labelSet) {
			selected = append(selected, publisher.PublisherID)
		}
	}
	sort.Ints(selected)
	return selected
}

func appMatchesTags(tags []privateAppTag, desired []string, mode string) bool {
	actual := map[string]bool{}
	for _, tag := range tags {
		actual[tag.TagName] = true
	}
	if mode == "all" {
		for _, tag := range desired {
			if !actual[tag] {
				return false
			}
		}
		return len(desired) > 0
	}
	for _, tag := range desired {
		if actual[tag] {
			return true
		}
	}
	return false
}

func currentPublisherIDs(assignments []privateAppPublisherAssignment) []int {
	ids := make([]int, 0, len(assignments))
	for _, assignment := range assignments {
		ids = append(ids, assignment.PublisherID)
	}
	sort.Ints(ids)
	return ids
}

func reconcilePublisherIDs(current []int, selected map[int]bool, matches bool) []int {
	nextSet := intSet(current)
	for id := range selected {
		if matches {
			nextSet[id] = true
		} else {
			delete(nextSet, id)
		}
	}
	return sortedIntSet(nextSet)
}

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func intSet(values []int) map[int]bool {
	set := make(map[int]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func intersectsStringSet(values []string, set map[string]bool) bool {
	for _, value := range values {
		if set[value] {
			return true
		}
	}
	return false
}

func sortedIntSet(set map[int]bool) []int {
	values := make([]int, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Ints(values)
	return values
}

func sameInts(left []int, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
