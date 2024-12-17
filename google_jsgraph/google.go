// package google_jsgraph provides google javascript bar graph and pie graph.
package google_jsgraph

import (
	"html/template"
	"regexp"
	"sort"
	"strings"

	"github.com/keep94/toolbox/http_util"
)

var (
	namePattern = regexp.MustCompile(`^[a-z0-9]+$`)
)

var (
	kGoogleGraphTemplateSpec = `
<script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
<script type="text/javascript">
  google.charts.load("current", {packages:[{{.Packages}}]});
  google.charts.setOnLoadCallback(drawCharts);
  function drawCharts() {
{{.Code}}
  }
</script>
`
)

var (
	kGoogleGraphTemplate = template.Must(template.New("googleJsGraph").Parse(kGoogleGraphTemplateSpec))
)

// GraphData represents a dataset to be graphed.
type GraphData interface {

	// The number of X data points
	XLen() int

	// The number of Y data points
	YLen() int

	// The title of the X labels
	XTitle() string

	// Return 0-based label for X axis
	XLabel(x int) string

	// Return 0-based label for Y axis
	YLabel(y int) string

	// Return value at (x, y)
	Value(x, y int) float64
}

// Graph represents a Google javascript graph
type Graph interface {

	// Emits the package names this graph depends on to the packages map.
	EmitPackages(packages map[string]struct{})

	// Emits the the code within the drawCharts() function that draws this
	// graph.
	EmitCode(name string, sb *strings.Builder)
}

// MustEmit emits the javascript chunk that renders the graphs.
// In graphs, the keys are the ids of the div tags where the graphs go.
// The keys must match [a-z0-9]+ or else MustEmit panics. The return value
// of MustEmit belongs in the head section of the html document.
func MustEmit(graphs map[string]Graph) template.HTML {
	if len(graphs) == 0 {
		return ""
	}
	names := make([]string, 0, len(graphs))
	for n := range graphs {
		names = append(names, n)
	}
	sort.Strings(names)

	var code strings.Builder
	packages := make(map[string]struct{})
	for _, name := range names {
		graphs[name].EmitPackages(packages)
	}
	for _, name := range names {
		if !isValidName(name) {
			panic("Names must match [a-z0-9]+")
		}
		graphs[name].EmitCode(name, &code)
	}
	v := &view{
		Packages: packagesAsString(packages),
		Code:     template.JS(code.String()),
	}
	var sb strings.Builder
	http_util.WriteTemplate(&sb, kGoogleGraphTemplate, v)
	return template.HTML(sb.String())
}

type view struct {
	Packages template.JS
	Code     template.JS
}

func packagesAsString(packages map[string]struct{}) template.JS {
	parts := make([]string, 0, len(packages))
	for name := range packages {
		parts = append(parts, name)
	}
	sort.Strings(parts)
	for i := range parts {
		parts[i] = "'" + parts[i] + "'"
	}
	return template.JS(strings.Join(parts, ", "))
}

func isValidName(name string) bool {
	return namePattern.MatchString(name)
}
