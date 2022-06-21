package chart

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AndreZiviani/aws-fuzzy/internal/peering"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	opentracing "github.com/opentracing/opentracing-go"
)

type VPCResult struct {
	VpcId   string       `json:"resourceId"`
	Tags    []ConfigTags `json:"tags"`
	OwnerId string       `json:"configuration.ownerId"`
}

type PeeringResult struct {
	Configuration PeeringConfiguration `json:"configuration"`
	Tags          []ConfigTags         `json:"tags"`
}

type PeeringConfiguration struct {
	RequesterVpc ConfigVpc `json:"requesterVpcInfo"`
	AccepterVpc  ConfigVpc `json:"accepterVpcInfo"`
	PeeringId    string    `json:"vpcPeeringConnectionId"`
}

type ConfigVpc struct {
	VpcId   string `json:"vpcId"`
	OwnerId string `json:"ownerId"`
	Region  string `json:"region"`
}

type ConfigTags struct {
	Value string `json:"value"`
	Key   string `json:"key"`
}

type PeeringConnection struct {
	Requester    string
	RequesterVpc string
	Accepter     string
	AccepterVpc  string
}

type Node struct {
	Id      string
	Name    string
	Account string
}

func (p *Peering) Execute(args []string) error {
	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "chart")
	defer span.Finish()

	peeringsJson, vpcsJson, err := peering.Peering(ctx, p.Profile, p.Account, p.Region)

	if err != nil {
		return err
	}

	tmp := strings.Join(peeringsJson[:], ",")
	tmp = fmt.Sprintf("[%s]", tmp)

	peerings := []PeeringResult{}
	_ = json.Unmarshal([]byte(tmp), &peerings)

	tmp = strings.Join(vpcsJson[:], ",")
	tmp = fmt.Sprintf("[%s]", tmp)

	vpcs := []VPCResult{}
	_ = json.Unmarshal([]byte(tmp), &vpcs)

	graph := NewGraph()

	nodes, links, categories := mapResult(peerings, vpcs)

	AddToGraph(graph, nodes, links, categories)

	page := components.NewPage()

	page.AddCharts(graph)
	f, err := os.Create("graph.html")
	if err != nil {
		panic(err)
	}

	page.Render(io.MultiWriter(f))
	return nil
}

func NewGraph() *charts.Graph {

	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width: "100%", Height: "95vh",
			AssetsHost: "https://cdn.jsdelivr.net/npm/echarts@4/dist/", //use updated upstream js
		}),
		charts.WithTitleOpts(opts.Title{Title: "Peering Connections"}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Type:  "png",
					Title: "Download as PNG",
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{Show: true}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
	)

	return graph
}

func mapResult(peerings []PeeringResult, vpcs []VPCResult) ([]opts.GraphNode, []opts.GraphLink, []*opts.GraphCategory) {
	categories := make([]*opts.GraphCategory, 0)
	categories = append(categories, &opts.GraphCategory{}) // workaround bug

	links := make([]opts.GraphLink, 0)
	nodes := make([]opts.GraphNode, 0)

	dnodes := make(map[string]Node)

	login := sso.Login{}
	login.LoadProfiles()

	var requesterAccountName, accepterAccountName string

	for _, peering := range peerings {

		requesterAccount, err := login.GetProfileFromID(peering.Configuration.RequesterVpc.OwnerId)
		if err == nil {
			requesterAccountName = fmt.Sprintf("%s\n(%s)", requesterAccount.Name, peering.Configuration.RequesterVpc.Region)
		} else {
			requesterAccountName = fmt.Sprintf("%s\n(%s)", peering.Configuration.RequesterVpc.OwnerId, peering.Configuration.RequesterVpc.Region)
		}

		accepterAccount, err := login.GetProfileFromID(peering.Configuration.AccepterVpc.OwnerId)
		if err == nil {
			accepterAccountName = fmt.Sprintf("%s\n(%s)", accepterAccount.Name, peering.Configuration.AccepterVpc.Region)
		} else {
			accepterAccountName = fmt.Sprintf("%s\n(%s)", peering.Configuration.AccepterVpc.OwnerId, peering.Configuration.AccepterVpc.Region)
		}

		requesterVpcId := peering.Configuration.RequesterVpc.VpcId
		requesterVpcName := getVpcName(vpcs, requesterVpcId)

		accepterVpcId := peering.Configuration.AccepterVpc.VpcId
		accepterVpcName := getVpcName(vpcs, accepterVpcId)

		dnodes[requesterVpcId] = Node{Id: requesterVpcId, Account: requesterAccountName, Name: requesterVpcName}
		dnodes[accepterVpcId] = Node{Id: accepterVpcId, Account: accepterAccountName, Name: accepterVpcName}

		links = append(links, opts.GraphLink{
			Source: requesterVpcName,
			Target: accepterVpcName,
		})
	}

	for _, v := range dnodes {

		categories = append(categories,
			&opts.GraphCategory{
				Name:  v.Account,
				Label: &opts.Label{Show: true, Color: "auto"},
			})

		nodes = append(nodes,
			opts.GraphNode{
				Name:     v.Name,
				Category: len(categories) - 1,
			},
		)

	}

	return nodes, links, categories

}

func AddToGraph(graph *charts.Graph, nodes []opts.GraphNode, links []opts.GraphLink, categories []*opts.GraphCategory) {
	graph.AddSeries("graph", nodes, links).
		SetSeriesOptions(
			charts.WithGraphChartOpts(
				opts.GraphChart{
					Force:              &opts.GraphForce{Repulsion: 100},
					Layout:             "force",
					Roam:               true,
					Categories:         categories,
					FocusNodeAdjacency: true,
				}),
			charts.WithLabelOpts(opts.Label{Show: true, Position: "right", Color: "Black"}),
			charts.WithEmphasisOpts(opts.Emphasis{
				Label: &opts.Label{
					Show:     true,
					Color:    "black",
					Position: "right",
				},
			}),
		)
}

func getVpcName(vpcs []VPCResult, key string) string {
	for _, vpc := range vpcs {
		if vpc.VpcId == key {
			for _, tag := range vpc.Tags {
				if tag.Key == "Name" {
					return fmt.Sprintf("%s\n(%s)", tag.Value, key)
				}
			}
		}
	}
	return key
}

func getTagName(tags []ConfigTags, key string) string {
	for _, tag := range tags {
		if tag.Key == "Name" {
			return tag.Value
		}
	}
	return key
}
