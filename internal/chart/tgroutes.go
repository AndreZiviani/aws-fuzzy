package chart

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/AndreZiviani/aws-fuzzy/internal/common"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/AndreZiviani/aws-fuzzy/internal/vpc"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	opentracing "github.com/opentracing/opentracing-go"
)

func (p *TGroutesCommand) Execute(args []string) error {
	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "chart")
	defer span.Finish()

	creds, err := sso.GetCredentials(ctx, p.Profile, false)
	if err != nil {
		return nil
	}

	cfg, err := sso.NewAwsConfig(ctx, creds)
	if err != nil {
		return nil
	}

	ec2client := ec2.NewFromConfig(cfg)
	if err != nil {
		return nil
	}

	tables, _ := vpc.DescribeTransitGatewayRouteTables(ctx, ec2client)

	tableNode := make([]*opts.TreeData, 0) // root children

	//TODO: iterate tables
	for _, table := range tables {
		ipRanges := make([]*opts.TreeData, 0)

		for _, ipRange := range table.Routes {
			subnet := aws.ToString(ipRange.DestinationCidrBlock)
			//TODO: iterate attachments
			attachmentId := aws.ToString(ipRange.TransitGatewayAttachments[0].TransitGatewayAttachmentId)

			attachments, _ := vpc.GetTransitGatewayAttachmentsByAttachment(ctx, ec2client, attachmentId)

			//TODO: iterate attachments
			account, _, err := sso.GetAccount(aws.ToString(attachments[0].ResourceOwnerId))
			if err != nil {
				account = attachments[0].ResourceOwnerId
			}

			name := fmt.Sprintf("%s\n%s\n%s",
				common.GetEC2Tag(attachments[0].Tags, "Name", aws.ToString(attachments[0].TransitGatewayAttachmentId)),
				aws.ToString(account),
				aws.ToString(attachments[0].ResourceId),
			)

			destination := []*opts.TreeData{
				&opts.TreeData{
					Name: name,
				},
			}

			ipNode := &opts.TreeData{
				Name:     subnet,
				Children: destination,
			}

			ipRanges = append(ipRanges, ipNode)
		}
		tmp := &opts.TreeData{
			Name:     table.Name,
			Children: ipRanges,
		}
		tableNode = append(tableNode, tmp)
	}

	rootNode := []opts.TreeData{
		opts.TreeData{
			Children: tableNode,
		},
	}

	g := NewTree("Transit Gateway Route Tables")
	g.AddSeries("tree", rootNode).
		SetSeriesOptions(
			charts.WithTreeOpts(
				opts.TreeChart{
					Layout:           "orthogonal",
					Orient:           "LR",
					InitialTreeDepth: -1,
					Roam:             true,
					Leaves: &opts.TreeLeaves{
						Label: &opts.Label{Show: true, Position: "right", Color: "Black"},
					},
				},
			),
			charts.WithLabelOpts(
				opts.Label{
					Show:     true,
					Position: "top",
					Color:    "Black",
				},
			),
		)

	page := NewPage()
	page.AddCharts(g)
	f, err := os.Create("tgroutes.html")
	if err != nil {
		panic(err)
	}

	page.Render(io.MultiWriter(f))

	return err
}
