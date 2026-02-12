package chart

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sort"
	"sync"

	"github.com/AndreZiviani/aws-fuzzy/internal/common"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/AndreZiviani/aws-fuzzy/internal/vpc"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	opentracing "github.com/opentracing/opentracing-go"
)

func NewTGRoutes(profile, region string) *TGRoutes {
	tgr := TGRoutes{
		Profile: profile,
		Region:  region,
	}
	return &tgr
}

type tgAttachmentRoutine struct {
	Index       int
	Subnet      *string
	Attachments []ec2types.TransitGatewayAttachment
}

func waitTGChannel(c chan<- *tgAttachmentRoutine, wg *sync.WaitGroup) {
	wg.Wait()
	close(c)
}

func getTGAttachmentsRoutine(ctx context.Context, ec2client *ec2.Client, i int, subnet *string, attachmentId string, c chan<- *tgAttachmentRoutine, wg *sync.WaitGroup) {
	defer wg.Done()
	attachments, _ := vpc.GetTransitGatewayAttachmentsByAttachment(ctx, ec2client, attachmentId)
	c <- &tgAttachmentRoutine{
		Index:       i, // so we can keep the order later
		Subnet:      subnet,
		Attachments: attachments,
	}
}

type TreeDataIp struct {
	Name     string
	Children []*net.IPNet
}

func processTables(ctx context.Context, ec2client *ec2.Client, tables []*vpc.DescribeTransitGatewayRouteTablesOutput) ([]*opts.TreeData, error) {
	tableNode := make([]*opts.TreeData, 0)

	login := sso.Login{}
	login.LoadProfiles()

	for _, table := range tables {
		var ipRanges []*opts.TreeData
		ipRanges = make([]*opts.TreeData, 0) // cant know beforehand how many destinations
		nodeMapIp := make(map[string]*TreeDataIp)

		c := make(chan *tgAttachmentRoutine)
		var wg sync.WaitGroup

		for i, ipRange := range table.Routes {
			wg.Add(1)

			//TODO: iterate attachments
			attachmentId := aws.ToString(ipRange.TransitGatewayAttachments[0].TransitGatewayAttachmentId)

			// Describe attachments in parallel
			go getTGAttachmentsRoutine(ctx, ec2client, i, ipRange.DestinationCidrBlock, attachmentId, c, &wg)
		}

		// Wait for all routines to finish, then close channel to free main thread
		go waitTGChannel(c, &wg)

		for attachmentsRoutine := range c {
			attachments := attachmentsRoutine.Attachments

			subnet := aws.ToString(attachmentsRoutine.Subnet)

			//TODO: iterate attachments
			profile, err := login.GetProfileFromID(aws.ToString(attachments[0].ResourceOwnerId))
			account := *attachments[0].ResourceOwnerId
			if err == nil {
				account = profile.Name
			}

			name := fmt.Sprintf("%s\n%s\n%s",
				common.GetEC2Tag(attachments[0].Tags, "Name", aws.ToString(attachments[0].TransitGatewayAttachmentId)),
				account,
				aws.ToString(attachments[0].ResourceId),
			)

			resourceId := aws.ToString(attachments[0].ResourceId)
			node, ok := nodeMapIp[resourceId]
			if !ok {
				nodeMapIp[resourceId] = &TreeDataIp{
					Name:     name,
					Children: make([]*net.IPNet, 0),
				}
				node = nodeMapIp[resourceId]
			}
			_, netsubnet, _ := net.ParseCIDR(subnet)
			node.Children = append(node.Children, netsubnet)

		}

		for _, v := range nodeMapIp {
			sort.Slice(v.Children, func(i, j int) bool {
				return bytes.Compare(v.Children[i].IP, v.Children[j].IP) < 0
			})
			tmp := make([]*opts.TreeData, 0)
			for _, i := range v.Children {
				tmp = append(tmp, &opts.TreeData{Name: i.String()})
			}
			ipRanges = append(ipRanges, &opts.TreeData{Name: v.Name, Children: tmp})
		}

		tmp := &opts.TreeData{
			Name:     table.Name,
			Children: ipRanges,
		}
		tableNode = append(tableNode, tmp)
	}

	return tableNode, nil
}

func (p *TGRoutes) Execute(ctx context.Context) error {
	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer func() { _ = closer.Close() }()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "tgroute chart")
	defer span.Finish()

	login := sso.Login{Profile: p.Profile}

	creds, err := login.GetCredentials(ctx)
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

	tableNode, _ := processTables(ctx, ec2client, tables) // root children

	rootNode := []opts.TreeData{
		{
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

	return page.Render(io.MultiWriter(f))
}
