package ssm

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func testInstance(id, ip, name string) Instance {
	i := Instance{}
	i.InstanceId = aws.String(id)
	i.PrivateIpAddress = aws.String(ip)
	if name != "" {
		i.Tags = []ec2types.Tag{{Key: aws.String("Name"), Value: aws.String(name)}}
	}
	return i
}

func fixtures() []Instance {
	return []Instance{
		testInstance("i-0aaa", "10.0.1.5", "web-prod-1"),
		testInstance("i-0bbb", "10.0.2.9", "web-prod-2"),
		testInstance("i-0ccc", "10.0.3.4", "DB-Prod"),
		testInstance("i-0ddd", "10.0.4.7", ""),
		// Name tag deliberately collides with another instance's private IP.
		testInstance("i-0eee", "10.0.5.2", "10.0.1.5"),
	}
}

func ids(instances []Instance) []string {
	out := make([]string, 0, len(instances))
	for _, i := range instances {
		out = append(out, aws.ToString(i.InstanceId))
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if a[k] != b[k] {
			return false
		}
	}
	return true
}

func TestMatchInstances(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{"by instance id", "i-0bbb", []string{"i-0bbb"}},
		{"by private ip", "10.0.3.4", []string{"i-0ccc"}},
		{"by exact name", "web-prod-1", []string{"i-0aaa"}},
		{"by name case insensitive", "db-prod", []string{"i-0ccc"}},
		{"by name substring", "web-prod", []string{"i-0aaa", "i-0bbb"}},
		{"no match", "does-not-exist", []string{}},
		{"empty query matches nothing", "", []string{}},
		{"instance without name tag is skipped by name tiers", "prod", []string{"i-0aaa", "i-0bbb", "i-0ccc"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ids(matchInstances(fixtures(), tt.query))
			if !equal(got, tt.want) {
				t.Errorf("matchInstances(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

// An exact IP match on one instance must win over a Name-tag match on another,
// even though both tiers would hit for the same query.
func TestMatchInstancesTierPrecedence(t *testing.T) {
	got := ids(matchInstances(fixtures(), "10.0.1.5"))
	want := []string{"i-0aaa"}
	if !equal(got, want) {
		t.Errorf("matchInstances(ip that is also a Name tag) = %v, want %v", got, want)
	}
}
