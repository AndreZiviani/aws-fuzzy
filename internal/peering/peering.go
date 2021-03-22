package peering

import (
	"context"

	"github.com/AndreZiviani/aws-fuzzy/internal/config"
	opentracing "github.com/opentracing/opentracing-go"
	//"github.com/opentracing/opentracing-go/log"
)

func Peering(ctx context.Context, profile string, account string) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "peering")
	defer span.Finish()

	query := config.ConfigCommand{
		Profile: profile,
		Account: account,
		Pager:   false,
		Service: "EC2",
		Select: "configuration.requesterVpcInfo.ownerId" +
			", configuration.requesterVpcInfo.vpcId" +
			", configuration.accepterVpcInfo.vpcId" +
			", configuration.accepterVpcInfo.ownerId" +
			", configuration.vpcPeeringConnectionId" +
			", tags.key, tags.value",
		Filter: "",
		Limit:  0,
	}

	result, err := config.Config(ctx, &query, "VPCPeeringConnection%")

	return result, err
}

/*
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
*/

/*
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
*/
