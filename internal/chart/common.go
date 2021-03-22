package chart

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func NewTree(title string) *charts.Tree {

	graph := charts.NewTree()
	graph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width: "100%", Height: "95vh",
			AssetsHost: "https://cdn.jsdelivr.net/npm/echarts@4/dist/", //use updated upstream js
		}),
		charts.WithTitleOpts(opts.Title{Title: title}),
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
		charts.WithTooltipOpts(opts.Tooltip{Show: false}),
	)

	return graph
}

func NewPage() *components.Page {
	page := components.NewPage()
	page.SetPageOptions(
		components.WithInitializationOpts(opts.Initialization{
			AssetsHost: "https://cdn.jsdelivr.net/npm/echarts@4/dist/", //use updated upstream js
		}),
	)

	return page
}
