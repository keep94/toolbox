// package google_jsgraph provides google javascript bar graph and pie graph.
package google_jsgraph

import (
	"errors"
	"html/template"
	"io"
	"regexp"
	"sort"
	"strings"
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

	// Packages returns the Google javascript packages this graph depends on
	Packages() []string

	// WriteCode writes the code within the drawCharts() function that draws
	// this graph. name is the id of the div tag where this graph will go.
	// When Emit or MustEmit calls this, it provides a w that also implements
	// io.ByteWriter and io.StringWriter.
	WriteCode(name string, w io.Writer) error
}

// MustEmit works like Emit except that when Emit returns an error, MustEmit
// panics.
func MustEmit(graphs map[string]Graph) template.HTML {
	result, err := Emit(graphs)
	if err != nil {
		panic(err)
	}
	return result
}

// Emit emits the javascript chunk that renders the graphs.
// In graphs, the keys are the ids of the div tags where the graphs go.
// The keys must match [a-z0-9]+ or else Emit returns an error. The
// return value of Emit belongs in the head section of the html document.
func Emit(graphs map[string]Graph) (template.HTML, error) {
	if len(graphs) == 0 {
		return "", nil
	}
	names := make([]string, 0, len(graphs))
	for n := range graphs {
		names = append(names, n)
	}
	sort.Strings(names)

	var code strings.Builder
	opqCode := &opqBuilder{delegate: &code}
	packages := make(map[string]struct{})
	for _, name := range names {
		for _, p := range graphs[name].Packages() {
			packages[p] = struct{}{}
		}
	}
	for _, name := range names {
		if !isValidName(name) {
			return "", errors.New("Names must match [a-z0-9]+")
		}
		if err := graphs[name].WriteCode(name, opqCode); err != nil {
			return "", err
		}
	}
	v := &view{
		Packages: packagesAsString(packages),
		Code:     template.JS(code.String()),
	}
	var sb strings.Builder
	if err := kGoogleGraphTemplate.Execute(&sb, v); err != nil {
		return "", err
	}
	return template.HTML(sb.String()), nil
}

type opqBuilder struct {
	delegate *strings.Builder
}

func (b *opqBuilder) Write(p []byte) (int, error) {
	return b.delegate.Write(p)
}

func (b *opqBuilder) WriteByte(c byte) error {
	return b.delegate.WriteByte(c)
}

func (b *opqBuilder) WriteString(s string) (int, error) {
	return b.delegate.WriteString(s)
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
