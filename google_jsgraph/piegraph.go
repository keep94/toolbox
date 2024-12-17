package google_jsgraph

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/keep94/toolbox/http_util"
)

var (
	kPieGraphTemplateSpec = `
var {{.DataVar}} = google.visualization.arrayToDataTable({{.Data}});
var {{.OptionsVar}} = {
  legend: "none",
  is3D: true,
  pieSliceText: "none",
  slices: {{.Colors}}
};
var {{.ChartVar}} = new google.visualization.PieChart(document.getElementById("{{.Name}}"))
{{.ChartVar}}.draw({{.DataVar}}, {{.OptionsVar}})
`
)

var (
	kPieGraphTemplate = template.Must(template.New("pieGraph").Parse(kPieGraphTemplateSpec))
)

// PieGraph represents a pie graph.
type PieGraph struct {

	// The graph data. YLen should be 1.
	Data GraphData

	// Optional: Palette consists of the RGB colors to use in the pie graph.
	// e.g []String{"FF0000", "00FF00", "0000FF"}. If omitted, Google chooses
	// the palette.
	Palette []string
}

func (p *PieGraph) EmitPackages(packages map[string]struct{}) {
	packages["corechart"] = struct{}{}
}

func (p *PieGraph) EmitCode(name string, sb *strings.Builder) {
	v := &pieview{
		Data:       asJSArray(p.Data),
		DataVar:    "data_" + name,
		OptionsVar: "options_" + name,
		ChartVar:   "chart_" + name,
		Name:       name,
		Colors:     p.paletteString(),
	}
	http_util.WriteTextTemplate(sb, kPieGraphTemplate, v)
}

func (p *PieGraph) paletteString() string {
	if len(p.Palette) == 0 {
		return "{\n}"
	}
	parts := make([]string, p.Data.XLen())
	for i := range parts {
		parts[i] = fmt.Sprintf(
			"%d: { color: %s }",
			i,
			quoteString("#"+p.Palette[i%len(p.Palette)]))
	}
	return "{\n" + strings.Join(parts, ",\n") + "\n}"
}

type pieview struct {
	Data       string
	DataVar    string
	OptionsVar string
	Colors     string
	ChartVar   string
	Name       string
}
