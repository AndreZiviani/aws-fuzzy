package ssm

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/AndreZiviani/aws-fuzzy/internal/common"
	"github.com/aws/aws-sdk-go-v2/aws"
	opentracing "github.com/opentracing/opentracing-go"
)

var instanceIDPattern = regexp.MustCompile(`^i-[0-9a-f]+$`)

// matchInstances filters instances by query, trying increasingly loose tiers
// and returning the first tier that has any hits: exact instance id, exact
// private IP, exact Name tag (case-insensitive), substring of the Name tag
// (case-insensitive). An empty query matches nothing.
func matchInstances(instances []Instance, query string) []Instance {
	if query == "" {
		return []Instance{}
	}

	lower := strings.ToLower(query)

	tiers := []func(Instance) bool{
		func(i Instance) bool { return aws.ToString(i.InstanceId) == query },
		func(i Instance) bool { return aws.ToString(i.PrivateIpAddress) == query },
		func(i Instance) bool { return strings.ToLower(instanceName(i)) == lower },
		func(i Instance) bool { return strings.Contains(strings.ToLower(instanceName(i)), lower) },
	}

	for _, matches := range tiers {
		hits := make([]Instance, 0)
		for _, i := range instances {
			if matches(i) {
				hits = append(hits, i)
			}
		}
		if len(hits) > 0 {
			return hits
		}
	}

	return []Instance{}
}

func instanceName(i Instance) string {
	return common.GetEC2Tag(i.Tags, "Name", "")
}

// validateTarget rejects a flag combination that could only be satisfied by
// opening the picker. Called before any AWS call so that a scripted run fails
// on the flags rather than on a missing login.
func validateTarget(query string, noTui bool) error {
	if query == "" && noTui {
		return fmt.Errorf("no instance specified, --instance is required with --non-interactive")
	}

	return nil
}

// resolveInstance turns a user-supplied target into a single instance.
//
// An empty query opens the fuzzy picker over every SSM-managed instance. A
// query that resolves to exactly one instance is returned directly; one that
// resolves to several opens the picker seeded with just those candidates.
// With noTui set, anything that would open the picker is an error instead, so
// the command can never block in a script.
func resolveInstance(ctx context.Context, cfg aws.Config, query string, noTui bool) (*Instance, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssmresolveinstance")
	defer span.Finish()

	if err := validateTarget(query, noTui); err != nil {
		return nil, err
	}

	// when we already know the exact instance id we can ask SSM about just
	// that one instance instead of enumerating the whole account
	var filter []string
	if instanceIDPattern.MatchString(query) {
		filter = []string{query}
	}

	instancesOutput, err := GetInstances(ctx, cfg, filter)
	if err != nil {
		return nil, err
	}

	instances := flattenInstances(instancesOutput)

	if query == "" {
		return tui(instances)
	}

	candidates := matchInstances(instances, query)

	switch {
	case len(candidates) == 0:
		return nil, fmt.Errorf("no running SSM-managed instance matches %q", query)
	case len(candidates) == 1:
		return &candidates[0], nil
	case noTui:
		return nil, fmt.Errorf("%q matches %d instances, re-run with one of:\n%s",
			query, len(candidates), describeCandidates(candidates))
	default:
		fmt.Fprintf(os.Stderr, "%d instances match %q, pick one:\n", len(candidates), query)
		return tui(candidates)
	}
}

func describeCandidates(candidates []Instance) string {
	out := strings.Builder{}
	for _, i := range candidates {
		name := instanceName(i)
		if name == "" {
			name = "<missing name>"
		}
		fmt.Fprintf(&out, "  %s  %s (%s)\n",
			aws.ToString(i.InstanceId), name, aws.ToString(i.PrivateIpAddress))
	}
	return strings.TrimRight(out.String(), "\n")
}
