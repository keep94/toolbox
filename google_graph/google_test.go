package google_graph

import (
  "net/url"
  "reflect"
  "testing"
)

func TestBarGraphEncodeData(t *testing.T) {
  bg := BarGraph{Palette: []string{"1"}, Scale: 2}
  gd := graphData{{"a", 30}, {"b", 312}}
  query := bg.GraphURL(gd).Query()
  verify(t, "s:Fw", query.Get("chd"))
  verify(t, "1,0,4", query.Get("chxr"))
}

func TestBarGraphEncodeData2D(t *testing.T) {
  bg := BarGraph{Palette: []string{"1", "2"}}
  gd := graphData2D{{"a", 30, 50}, {"b", 75, -4}, {"c", 50, 20}}
  query := bg.GraphURL2D(gd).Query()
  verify(t, "s:Y9p,pAQ", query.Get("chd"))
  verify(t, "1,0,75", query.Get("chxr"))
}

func TestBarGraphHideTitlesIfAllEmpty(t *testing.T) {
  bg := BarGraph{Palette: []string{"1", "2"}}
  gd := graphData2D{{"a", 30, 50}, {"b", 75, -4}, {"c", 50, 20}}
  query := bg.GraphURL2D(gd).Query()
  _, ok := query["chdl"]
  if ok {
    t.Error("Did not expect chdl parameter when no titles present.")
  }
}

func TestBarGraphZero(t *testing.T) {
  bg := BarGraph{Palette: []string{"1", "2"}}
  gd := graphData2D{{"a", -5, 0}, {"b", -3, -4}, {"c", 0, 0}}
  query := bg.GraphURL2D(gd).Query()
  verify(t, "s:AAA,AAA", query.Get("chd"))
  verify(t, "1,0,1", query.Get("chxr"))
}

func TestNoBarGraph(t *testing.T) {
  bg := BarGraph{Palette: []string{"1", "2"}}
  gd := graphData2D{}
  url := bg.GraphURL2D(gd)
  if url != nil {
    t.Error("Expect no graph URL for empty dataset.")
  }
}

func TestBarGraph(t *testing.T) {
  bg := BarGraph{Palette: []string{"FF0000", "00FF00"}}
  gd := withTitle{graphData2D{{"a", 30, 50}, {"b", 75, -4}, {"c", 50, 20}}}
  expected, _ := url.Parse("http://chart.apis.google.com/chart?chs=500x250&cht=bvg&chco=FF0000%2C00FF00&chd=s:Y9p,pAQ&chxl=0:|a|b|c&chxt=x,y&chxr=1,0,75&chbh=a&chdl=Income%7CExpense")
  actual := bg.GraphURL2D(gd)
  verifyUrl(t, expected, actual)
}

func TestPieGraphEncodeColors(t *testing.T) {
  pg := PieGraph{Palette: []string{"1", "2", "3"}}
  gd := graphData{{"a", 0}, {"b", 0}, {"c", 0}, {"d", 0}}
  verify(t, "1|2|3|1", pg.GraphURL(gd).Query().Get("chco"))
  gd = graphData{{"a", 0}, {"b", 0}, {"c", 0}}
  verify(t, "1|2|3", pg.GraphURL(gd).Query().Get("chco"))
  gd = graphData{{"a", 0}}
  verify(t, "1", pg.GraphURL(gd).Query().Get("chco"))
}

func TestPieGraphEncodeData(t *testing.T) {
  pg := PieGraph{Palette: []string{"1", "2", "3"}}
  gd := graphData{{"a", -5}, {"b", -3}}
  verify(t, "s:AA", pg.GraphURL(gd).Query().Get("chd"))
  gd = graphData{{"a", 0}, {"b", -3}}
  verify(t, "s:AA", pg.GraphURL(gd).Query().Get("chd"))
  gd = graphData{{"a", 10}, {"b", 15}, {"c", -5}}
  verify(t, "s:p9A", pg.GraphURL(gd).Query().Get("chd"))
}

func TestPieGraph(t *testing.T) {
  data := graphData{{"a", 10}, {"b", 15}, {"c", -5}}
  pg := PieGraph{Palette: []string{"FF0000", "00FF00"}}
  expected, _ := url.Parse("http://chart.apis.google.com/chart?chs=500x250&cht=p3&chco=FF0000%7C00FF00%7CFF0000&chd=s:p9A&chdl=a%7Cb%7Cc")
  actual := pg.GraphURL(data)
  verifyUrl(t, expected, actual)
}

func TestPieGraphEmptyDataset(t *testing.T) {
  pg := PieGraph{Palette: []string{"FF0000", "00FF00"}}
  url := pg.GraphURL(graphData{})
  if url != nil {
    t.Error("Expect no graph URL for empty dataset.")
  }
}

func verify(t *testing.T, expected, actual string) {
  if expected != actual {
    t.Errorf("Expected %s, got %s", expected, actual)
  }
}

func verifyUrl(t *testing.T, expected, actual *url.URL) {
  verify(t, expected.Scheme, actual.Scheme)
  verify(t, expected.Host, actual.Host)
  verify(t, expected.Path, actual.Path)
  if !reflect.DeepEqual(expected.Query(), actual.Query()) {
    t.Errorf("Expected %v, got %v", expected.Query(), actual.Query())
  }
}

type graphItem struct {
  Label string
  Value int64
}

type graphData []graphItem

func (g graphData) Len() int { return len(g) }
func (g graphData) Label(idx int) string { return g[idx].Label }
func (g graphData) Title() string { return "" }
func (g graphData) Value(idx int) int64 { return g[idx].Value }

type graphItem2D struct {
  Label string
  Val1 int64
  Val2 int64
}

type graphData2D []graphItem2D

func (g graphData2D) XLen() int { return len(g) }
func (g graphData2D) YLen() int { return 2 }
func (g graphData2D) XLabel(idx int) string { return g[idx].Label }
func (g graphData2D) YLabel(idx int) string { return "" }
func (g graphData2D) Value(x, y int) int64 {
  if y == 1 {
     return g[x].Val2
  }
  return g[x].Val1
}

type withTitle struct {
  graphData2D
}

func (g withTitle) YLabel(idx int) string {
  if idx == 0 {
    return "Income"
  }
  return "Expense"
}

