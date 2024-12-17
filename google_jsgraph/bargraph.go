package google_jsgraph

import (
	"strings"
	"text/template"

	"github.com/keep94/toolbox/http_util"
)

var (
	kBarGraphTemplateSpec = `
var {{.DataVar}} = google.visualization.arrayToDataTable({{.Data}});
var {{.OptionsVar}} = {
  legend: { position: "none" },
  bars: "vertical",
  vAxis: {format: "decimal"},
  colors: {{.Colors}}
};
var {{.ChartVar}} = new google.charts.Bar(document.getElementById("{{.Name}}"))
{{.ChartVar}}.draw({{.DataVar}}, google.charts.Bar.convertOptions({{.OptionsVar}}))
`
)

var (
	kBarGraphTemplate = template.Must(template.New("barGraph").Parse(kBarGraphTemplateSpec))
)

// BarGraph represents a bar graph.
type BarGraph struct {

	// The graph data
	Data GraphData

	// Palette consists of the RGB colors to use in the bar graph.
	// e.g []String{"FF0000", "00FF00", "0000FF"}
	Palette []string
}

func (b *BarGraph) EmitPackages(packages map[string]struct{}) {
	packages["bar"] = struct{}{}
}

func (b *BarGraph) EmitCode(name string, sb *strings.Builder) {
	v := &barview{
		Data:       asJSArray(b.Data),
		DataVar:    "data_" + name,
		OptionsVar: "options_" + name,
		ChartVar:   "chart_" + name,
		Name:       name,
		Colors:     b.paletteString(),
	}
	http_util.WriteTextTemplate(sb, kBarGraphTemplate, v)
}

func (b *BarGraph) paletteString() string {
	parts := make([]string, 0, len(b.Palette))
	for _, c := range b.Palette {
		parts = append(parts, quoteString("#"+c))
	}
	return asList(parts)
}

type barview struct {
	Data       string
	DataVar    string
	OptionsVar string
	Colors     string
	ChartVar   string
	Name       string
}
