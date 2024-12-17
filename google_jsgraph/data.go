package google_jsgraph

import (
	"strconv"
	"strings"
	"text/template"
)

func asJSArray(gd GraphData) string {
	parts := make([]string, 0, gd.XLen()+1)
	parts = append(parts, dataHeading(gd))
	for i := 0; i < gd.XLen(); i++ {
		parts = append(parts, dataRow(gd, i))
	}
	return "[\n" + strings.Join(parts, ",\n") + "\n]"
}

func dataHeading(gd GraphData) string {
	parts := make([]string, 0, gd.YLen()+1)
	parts = append(parts, quoteString(gd.XTitle()))
	for i := 0; i < gd.YLen(); i++ {
		parts = append(parts, quoteString(gd.YLabel(i)))
	}
	return asList(parts)
}

func dataRow(gd GraphData, row int) string {
	parts := make([]string, 0, gd.YLen()+1)
	parts = append(parts, quoteString(gd.XLabel(row)))
	for i := 0; i < gd.YLen(); i++ {
		parts = append(
			parts, strconv.FormatFloat(gd.Value(row, i), 'g', -1, 64))
	}
	return asList(parts)
}

func quoteString(s string) string {
	return "\"" + template.JSEscapeString(s) + "\""
}

func asList(parts []string) string {
	return "[" + strings.Join(parts, ", ") + "]"
}
