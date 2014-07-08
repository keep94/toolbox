// package google_graph provides google bar graph and pie graph.
package google_graph

import (
  "fmt"
  "github.com/keep94/appcommon/http_util"
  "net/url"
  "strings"
)

const (
  kGoogleAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

// GraphData represents a dataset to be graphed.
type GraphData interface {
  // The number of data points.
  Len() int
  // The title
  Title() string
  // The label of the 0-based idx data point.
  Label(idx int) string
  // The value of the 0-based idx data point.
  Value(idx int) int64
}

// GraphData2D represents a 2D dataset to be graphed.
type GraphData2D interface {
  // The number of X data points
  XLen() int
  // The number of Y data points
  YLen() int
  // Return 0-based label for X axis
  XLabel(x int) string
  // Return 0-based label for Y axis
  YLabel(y int) string
  // Return value at (x, y)
  Value(x, y int) int64
}

// Grapher returns the URL for a graph of a dataset.
type Grapher interface {
  GraphURL(gd GraphData) *url.URL
}

// Grapher2D returns the URL for a graph of a 2D dataset.
type Grapher2D interface {
  GraphURL2D(gd GraphData2D) *url.URL
}

// BarGraph builds a link to a google bar graph.
type BarGraph struct {
  // Palette consists of the RGB colors to use in the bar graph.
  // e.g []String{"FF0000", "00FF00", "0000FF"}
  Palette []string
  // a value of 10^Scale is one unit on bar graph.
  Scale int
}

// GraphURL returns a link to a bar graph displaying particular graph data.
// GraphURL returns nil if given graph data of length 0.
func (b *BarGraph) GraphURL(gd GraphData) *url.URL {
  return b.GraphURL2D(to2D{gd})
}

// GraphURL2D returns a link to a bar graph displaying particular graph data.
// GraphURL2D returns nil if given graph data of length 0 in either dimension.
func (b *BarGraph) GraphURL2D(gd GraphData2D) *url.URL {
  xlength := gd.XLen()
  ylength := gd.YLen()
  if xlength <= 0 || ylength <= 0 {
    return nil
  }
  labels := make([]string, xlength)
  titles := make([]string, ylength)
  values := make([][]int64, ylength)
  var includeChdl bool
  for y := range values {
    titles[y] = gd.YLabel(y)
    if titles[y] != "" {
      includeChdl = true
    }
    values[y] = make([]int64, xlength)
  }
  for x := range labels {
    labels[x] = gd.XLabel(x)
    for y := range values {
      values[y][x] = gd.Value(x, y)
    }
  }
  max := maxInt64(values...)
  if max == 0 {
    max = 1
  }
  for i := 0; i < b.Scale; i++ {
    max = (max + 9) / 10
  }
  actualMax := max
  for i := 0; i < b.Scale; i++ {
    max *= 10
  }

  encoded := encodeInt64(max, values...)
  url, _ := url.Parse("http://chart.apis.google.com/chart")
  urlParams := []string {
      "chs", "500x250",
      "cht", "bvg",
      "chco", encodeColors(len(values), b.Palette, ","),
      "chd", encoded,
      "chxt", "x,y",
      "chbh", "a",
      "chxr", fmt.Sprintf("1,0,%d", actualMax),
      "chxl", fmt.Sprintf("0:|%s", strings.Join(labels, "|")),
      "chdl", strings.Join(titles, "|")}
  // If we aren't including chdl parameter, chop it off of end of parameter
  // list
  if !includeChdl {
    urlParams = urlParams[:len(urlParams) - 2]
  }
  return http_util.AppendParams(url, urlParams...)
}

// PieGraph builds a link to a google pie graph.
type PieGraph struct {
  // Palette consists of the RGB colors to use in the pie graph.
  // e.g []String{"FF0000", "00FF00", "0000FF"}
  Palette []string
}

// GraphURL returns a link to a pie graph displaying particular graph data.
// GraphURL returns nil if given graph data of length 0.
func (p *PieGraph) GraphURL(gd GraphData) *url.URL {
  length := gd.Len()
  if length <= 0 {
    return nil
  }
  labels := make([]string, length)
  values := make([]int64, length)
  for idx := range labels {
    labels[idx] = gd.Label(idx)
    values[idx] = gd.Value(idx)
  }
  encoded := encodeInt64(maxInt64(values), values)
  url, _ := url.Parse("http://chart.apis.google.com/chart")
  return http_util.AppendParams(
      url,
      "chs", "500x250",
      "cht", "p3",
      "chco", encodeColors(len(values), p.Palette, "|"),
      "chd", encoded,
      "chdl", strings.Join(labels, "|"))
}

type to2D struct {
  GraphData
}

func (t to2D) XLen() int {
  return t.Len()
}

func (t to2D) YLen() int {
  return 1
}

func (t to2D) XLabel(x int) string {
  return t.Label(x)
}

func (t to2D) YLabel(x int) string {
  return t.Title()
}

func (t to2D) Value(x, y int) int64 {
  return t.GraphData.Value(x)
}

func encodeInt64(max int64, datasets ...[]int64) string {
  encoded := make([]string, len(datasets))
  for idx := range datasets {
    encoded[idx] = _encodeInt64(datasets[idx], max)
  }
  return fmt.Sprintf("s:%s", strings.Join(encoded, ","))
}

func _encodeInt64(data []int64, max int64) string {
  buffer := make([]byte, len(data))
  for idx := range data {
    buffer[idx] = kGoogleAlphabet[scaleInt64For61(data[idx], max)]
  }
  return string(buffer)
}

func scaleInt64For61(amount, max int64) int64 {
  if amount <= 0 {
    return 0
  }
  return (amount * 61 + max / 2) / max
}

func encodeColors(count int, palette []string, separator string) string {
  colors := make([]string, count)
  plen := len(palette)
  for idx := range colors {
    colors[idx] = palette[idx % plen]
  }
  return strings.Join(colors, separator)
}

func maxInt64(data ...[]int64) int64 {
  var result int64
  for _, v1 := range data {
    for _, v2 := range v1 {
      if v2 > result {
        result = v2
      }
    }
  }
  return result
}
