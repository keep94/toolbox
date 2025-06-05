package google_jsgraph

import (
	"errors"
	"io"
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

func TestMustEmitPanicsFromError(t *testing.T) {
	assert.Panics(t, func() {
		MustEmit(map[string]Graph{
			"error": errorGraphForTesting{},
		})
	})
}

func TestMustEmitEmpty(t *testing.T) {
	assert.Empty(t, MustEmit(nil))
}

func TestTypeAssertionFails(t *testing.T) {
	expected := `
<script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
<script type="text/javascript">
  google.charts.load("current", {packages:[]});
  google.charts.setOnLoadCallback(drawCharts);
  function drawCharts() {
Builder Failure

ByteWriter Success

StringWriter Success

A
  }
</script>
`
	chunk := MustEmit(map[string]Graph{
		"typeassertiongraph": typeAssertionGraph{},
	})
	assert.Equal(t, expected, string(chunk))
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
	pg.WriteCode("piegraph", &sb)
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

type errorGraphForTesting struct {
}

func (b errorGraphForTesting) Packages() []string {
	return nil
}

func (b errorGraphForTesting) WriteCode(name string, w io.Writer) error {
	return errors.New("Error!")
}

type barGraphForTesting struct {
}

func (b barGraphForTesting) Packages() []string {
	return []string{"bar", "baz"}
}

func (b barGraphForTesting) WriteCode(name string, w io.Writer) error {
	_, err := io.WriteString(w, "Bar graph code\n\n")
	return err
}

type pieGraphForTesting struct {
}

func (p pieGraphForTesting) Packages() []string {
	return []string{"foo", "bar"}
}

func (p pieGraphForTesting) WriteCode(name string, w io.Writer) error {
	_, err := io.WriteString(w, "Pie graph code\n\n")
	return err
}

type typeAssertionGraph struct {
}

func (t typeAssertionGraph) Packages() []string {
	return nil
}

func (t typeAssertionGraph) WriteCode(name string, w io.Writer) error {
	_, ok := w.(*strings.Builder)
	if ok {
		if _, err := io.WriteString(w, "Builder Success\n\n"); err != nil {
			return err
		}
	} else {
		if _, err := io.WriteString(w, "Builder Failure\n\n"); err != nil {
			return err
		}
	}

	_, ok = w.(io.ByteWriter)
	if ok {
		if _, err := io.WriteString(w, "ByteWriter Success\n\n"); err != nil {
			return err
		}
	} else {
		if _, err := io.WriteString(w, "ByteWriter Failure\n\n"); err != nil {
			return err
		}
	}

	_, ok = w.(io.StringWriter)
	if ok {
		if _, err := io.WriteString(w, "StringWriter Success\n\n"); err != nil {
			return err
		}
	} else {
		if _, err := io.WriteString(w, "StringWriter Failure\n\n"); err != nil {
			return err
		}
	}
	bw := w.(io.ByteWriter)
	return bw.WriteByte(0x41) // A
}
