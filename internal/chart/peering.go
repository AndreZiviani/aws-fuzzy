package chart

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AndreZiviani/aws-fuzzy/internal/config"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	opentracing "github.com/opentracing/opentracing-go"
	//"github.com/opentracing/opentracing-go/log"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type ConfigResult struct {
	Configuration ConfigConfiguration `json:"configuration"`
	Tags          []ConfigTags        `json:"tags"`
}

type ConfigConfiguration struct {
	RequesterVpc ConfigVpc `json:"requesterVpcInfo"`
	AccepterVpc  ConfigVpc `json:"accepterVpcInfo"`
	PeeringId    string    `json:"vpcPeeringConnectionId"`
}

type ConfigVpc struct {
	VpcId   string `json:"vpcId"`
	OwnerId string `json:"ownerId"`
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

func NewGraph() *charts.Graph {

	graph := charts.NewGraph()
	graph.SetGlobalOptions(
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

func MapResult(result []ConfigResult) ([]opts.GraphNode, []opts.GraphLink, []*opts.GraphCategory) {
	categories := make([]*opts.GraphCategory, 0)
	categories = append(categories, &opts.GraphCategory{}) // workaround bug

	links := make([]opts.GraphLink, 0)
	nodes := make([]opts.GraphNode, 0)

	dnodes := make(map[string]string)

	for _, peering := range result {

		dnodes[peering.Configuration.RequesterVpc.VpcId] = peering.Configuration.RequesterVpc.OwnerId
		dnodes[peering.Configuration.AccepterVpc.VpcId] = peering.Configuration.AccepterVpc.OwnerId

		links = append(links, opts.GraphLink{
			Source: peering.Configuration.RequesterVpc.VpcId,
			Target: peering.Configuration.AccepterVpc.VpcId,
			//Value:  peering.Configuration.PeeringId,
		})
	}

	for k, v := range dnodes {

		categories = append(categories,
			&opts.GraphCategory{
				Name:  v,
				Label: &opts.Label{Show: true, Color: "auto"},
			})

		nodes = append(nodes,
			opts.GraphNode{
				Name:     k,
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

func Peering(ctx context.Context, p *ChartCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "peering")
	defer span.Finish()
	query := config.ConfigCommand{
		Profile: p.Profile,
		Pager:   false,
		Service: "EC2",
		Select: "configuration.requesterVpcInfo.ownerId" +
			", configuration.requesterVpcInfo.vpcId" +
			", configuration.accepterVpcInfo.vpcId" +
			", configuration.accepterVpcInfo.ownerId" +
			", configuration.vpcPeeringConnectionId" +
			", tags.key, tags.value",
		Filter:  "",
		Limit:   0,
		Account: "",
	}

	result, _ := config.Config(ctx, &query, "VPCPeeringConnection%")
	tmp := strings.Join(result[:], ",")
	tmp = fmt.Sprintf("[%s]", tmp)

	o := []ConfigResult{}
	_ = json.Unmarshal([]byte(tmp), &o)

	//////////////////// graph

	graph := NewGraph()

	nodes, links, categories := MapResult(o)

	AddToGraph(graph, nodes, links, categories)

	page := components.NewPage()
	page.AddCharts(graph)
	f, err := os.Create("graph.html")
	if err != nil {
		panic(err)
	}

	page.Render(io.MultiWriter(f))
	//config.Print(false, r)
	return nil
}

func (p *ChartCommand) Execute(args []string) error {
	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "chart")
	defer span.Finish()

	return Peering(ctx, p)

}