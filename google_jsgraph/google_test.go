package google_jsgraph

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustEmit(t *testing.T) {
	expected := `
<script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
<script type="text/javascript">
  google.charts.load("current", {packages:['bar', 'baz', 'foo']});
  google.charts.setOnLoadCallback(drawCharts);
  function drawCharts() {
Bar graph code

Pie graph code


  }
</script>
`
	chunk := MustEmit(map[string]Graph{
		"bargraph": barGraphForTesting{},
		"piegraph": pieGraphForTesting{},
	})
	assert.Equal(t, expected, string(chunk))
}

func TestMustEmitPanics(t *testing.T) {
	assert.Panics(t, func() {
		MustEmit(map[string]Graph{
			"bar_graph": barGraphForTesting{},
		})
	})
}

func TestMustEmitEmpty(t *testing.T) {
	assert.Empty(t, MustEmit(nil))
}

func TestRealGraphs(t *testing.T) {
	expected := `
<script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
<script type="text/javascript">
  google.charts.load("current", {packages:['bar', 'corechart']});
  google.charts.setOnLoadCallback(drawCharts);
  function drawCharts() {

var data_bargraph = google.visualization.arrayToDataTable([
["title\"", "\"expense", "\"income"],
["\"01", 1.01, 2.02],
["\"02", 3.03, 4.04],
["\"03", 5.05, 6.06],
["\"04", 7.07, 8.08]
]);
var options_bargraph = {
  legend: { position: "none" },
  bars: "vertical",
  vAxis: {format: "decimal"},
  colors: ["#990000", "#006600"]
};
var chart_bargraph = new google.charts.Bar(document.getElementById("bargraph"))
chart_bargraph.draw(data_bargraph, google.charts.Bar.convertOptions(options_bargraph))

var data_piegraph = google.visualization.arrayToDataTable([
["Category", "Amount"],
["Car", 156.35],
["Bicycle", 28.52],
["Food", 59.36],
["Hobbies", 78.52]
]);
var options_piegraph = {
  legend: "none",
  is3D: true,
  pieSliceText: "none",
  slices: {
0: { color: "#000066" },
1: { color: "#666600" },
2: { color: "#660000" },
3: { color: "#000066" }
}
};
var chart_piegraph = new google.visualization.PieChart(document.getElementById("piegraph"))
chart_piegraph.draw(data_piegraph, options_piegraph)

  }
</script>
`
	bardata := &fakeGraphData{
		title:   "title\"",
		xlabels: []string{"\"01", "\"02", "\"03", "\"04"},
		ylabels: []string{"\"expense", "\"income"},
		values:  []float64{1.01, 2.02, 3.03, 4.04, 5.05, 6.06, 7.07, 8.08},
	}
	piedata := &fakeGraphData{
		title:   "Category",
		xlabels: []string{"Car", "Bicycle", "Food", "Hobbies"},
		ylabels: []string{"Amount"},
		values:  []float64{156.35, 28.52, 59.36, 78.52},
	}
	bg := &BarGraph{Data: bardata, Palette: []string{"990000", "006600"}}
	pg := &PieGraph{Data: piedata, Palette: []string{"000066", "666600", "660000"}}
	chunk := MustEmit(map[string]Graph{"bargraph": bg, "piegraph": pg})
	assert.Equal(t, expected, string(chunk))
}

func TestPieGraphNoPalette(t *testing.T) {
	expected := `
var data_piegraph = google.visualization.arrayToDataTable([
["Category", "Amount"],
["Car", 156.35],
["Bicycle", 28.52],
["Food", 59.36],
["Hobbies", 78.52]
]);
var options_piegraph = {
  legend: "none",
  is3D: true,
  pieSliceText: "none",
  slices: {
}
};
var chart_piegraph = new google.visualization.PieChart(document.getElementById("piegraph"))
chart_piegraph.draw(data_piegraph, options_piegraph)
`
	piedata := &fakeGraphData{
		title:   "Category",
		xlabels: []string{"Car", "Bicycle", "Food", "Hobbies"},
		ylabels: []string{"Amount"},
		values:  []float64{156.35, 28.52, 59.36, 78.52},
	}
	pg := &PieGraph{Data: piedata}
	var sb strings.Builder
	pg.EmitCode("piegraph", &sb)
	assert.Equal(t, expected, sb.String())
}

type fakeGraphData struct {
	title   string
	xlabels []string
	ylabels []string
	values  []float64
}

func (f *fakeGraphData) XLen() int           { return len(f.xlabels) }
func (f *fakeGraphData) YLen() int           { return len(f.ylabels) }
func (f *fakeGraphData) XTitle() string      { return f.title }
func (f *fakeGraphData) XLabel(x int) string { return f.xlabels[x] }
func (f *fakeGraphData) YLabel(y int) string { return f.ylabels[y] }
func (f *fakeGraphData) Value(x, y int) float64 {
	return f.values[x*f.YLen()+y]
}

type barGraphForTesting struct {
}

func (b barGraphForTesting) EmitPackages(packages map[string]struct{}) {
	packages["bar"] = struct{}{}
	packages["baz"] = struct{}{}
}

func (b barGraphForTesting) EmitCode(name string, sb *strings.Builder) {
	sb.WriteString("Bar graph code\n\n")
}

type pieGraphForTesting struct {
}

func (p pieGraphForTesting) EmitPackages(packages map[string]struct{}) {
	packages["foo"] = struct{}{}
	packages["bar"] = struct{}{}
}

func (p pieGraphForTesting) EmitCode(name string, sb *strings.Builder) {
	sb.WriteString("Pie graph code\n\n")
}
