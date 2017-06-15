// Package http_util provides useful routines for writing web apps.
package http_util

import (
  "bytes"
  "fmt"
  "github.com/keep94/gofunctional3/consume"
  "html/template"
  "io"
  "log"
  "net/http"
  "net/url"
  "os"
  "strconv"
  "time"
)

var (
  kLog *log.Logger
  kAppStart time.Time
)

// Redirect sends a 302 redirect
func Redirect(w http.ResponseWriter, r *http.Request, redirectUrl string) {
  http.Redirect(w, r, redirectUrl, 302)
}

// HasParam returns true if values contains a particular parameter.
func HasParam(values url.Values, param string) bool {
  _, ok := values[param]
  return ok
}

// WithParams returns a URL with new parameters. If the parameters
// already exist in the original URL, they are replaced.
// u is the original URL;
// nameValues is parameter name, parameter value, parameter name, parameter
// value, etc. nameValues must have even length.
func WithParams(u *url.URL, nameValues ...string) *url.URL {
  length := len(nameValues)
  if length % 2 != 0 {
    panic("nameValues must have even length.")
  }
  result := *u
  values := result.Query()
  for i := 0; i < length; i += 2 {
    values.Set(nameValues[i], nameValues[i + 1])
  }
  result.RawQuery = values.Encode()
  return &result
}

// NewUrl returns a new URL with a given path and parameters.
// nameValues is parameter name, parameter value, parameter name, parameter
// value, etc. nameValues must have even length.
func NewUrl(path string, nameValues ...string) *url.URL {
  length := len(nameValues)
  if length % 2 != 0 {
    panic("nameValues must have even length.")
  }
  values := make(url.Values)
  for i := 0; i < length; i += 2 {
    values.Add(nameValues[i], nameValues[i + 1])
  }
  return &url.URL{
      Path: path,
      RawQuery: values.Encode()}
}

// AppendParams returns a URL with new parameters appended. No existing
// parameter is replaced. u is the original URL; nameValues is
// parameter name, parameter value, parameter name, parameter
// value, etc. nameValues must have even length.
func AppendParams(u *url.URL, nameValues ...string) *url.URL {
  length := len(nameValues)
  if length % 2 != 0 {
    panic("nameValues must have even length.")
  }
  result := *u
  values := result.Query()
  for i := 0; i < length; i += 2 {
    values.Add(nameValues[i], nameValues[i + 1])
  }
  result.RawQuery = values.Encode()
  return &result
}

// Pager simplifies displaying data in a PageBuffer using go templates.
type Pager struct {
  // The PageBuffer
  *consume.PageBuffer
  // The current URL
  URL *url.URL
  // The page number URL parameter name
  PageNoParam string
}

// DisplayPageNo returns the 1-based page number.
func (p *Pager) DisplayPageNo() int {
  return p.PageNo() + 1
}

// NextPageLink returns the URL for the next page.
func (p *Pager) NextPageLink() *url.URL {
  return WithParams(p.URL, p.PageNoParam, strconv.Itoa(p.PageNo() + 1))
}

// PrevPageLink returns the URL for the previous page.
func (p *Pager) PrevPageLink() *url.URL {
  return WithParams(p.URL, p.PageNoParam, strconv.Itoa(p.PageNo() - 1))
}

// WriteTemplate writes a template. v is the values for the template.
func WriteTemplate(w io.Writer, t *template.Template, v interface{}) {
  if err := t.Execute(w, v); err != nil {
    fmt.Fprintln(w, "Error in template.")
    kLog.Printf("Error in template: %v\n", err)
  }
}

// Repoort error reports an error. message is what user sees.
func ReportError(w http.ResponseWriter, message string, err error) {
  http.Error(w, message, http.StatusInternalServerError)
  kLog.Printf("%s: %v\n", message, err)
}

// Selection represents a single selection from a drop down
type Selection struct {
  // Value is the value of the selection
  Value string
  // Name is what is displayed for the selection
  Name string
}

// SelectModel converts a parameter value to a selection.
type SelectModel interface {
  // ToSelection converts a parameter value to a selection. ToSelection may
  // return nil if value does not map to a valid selection.
  ToSelection(s string) *Selection
}

// Choice represents a choice in a combo box.
type Choice struct {
  // What the user sees in the choice dialog
  Name string
  // The parameter value attached to this choice
  Value interface{}
}

// ComboBox represents an immutable combo box of items.
// ComboBox implements SelectModel.
type ComboBox []Choice

func (c ComboBox) ToSelection(s string) *Selection {
  if idx, ok := c.toIdx(s); ok {
    return &Selection{Name: c[idx].Name, Value: s}
  }
  return nil
}

// ToValue returns the value associated with the selected choice or nil
// if none selected. s is the value from the form.
func (c ComboBox) ToValue(s string) interface{} {
  if idx, ok := c.toIdx(s); ok {
    return c[idx].Value
  }
  return nil
}

// Items returns all the items in this combo box.
func (c ComboBox) Items() []Selection {
  result := make([]Selection, len(c))
  for i := range c {
    result[i] = Selection{Name: c[i].Name, Value: strconv.Itoa(i + 1)}
  }
  return result
}

func (c ComboBox) toIdx(s string) (int, bool) {
  oneIdx, _ := strconv.Atoi(s)
  idx := oneIdx - 1
  if idx < 0 || idx >= len(c) {
    return 0, false
  }
  return idx, true
}

// Selections implements SelectModel
type Selections []Selection

func (s Selections) ToSelection(str string) *Selection {
  for _, sel := range s {
    if str == sel.Value {
      return &sel
    }
  }
  return nil
}

// Values is a wrapper around url.Values providing additional methods.
type Values struct {
  url.Values
}

// GetSelection gets the current selection. name is the request parameter
// name. GetSelection may return nil if there is no valid selection.
func (v Values) GetSelection(
    model SelectModel, name string) *Selection {
  return model.ToSelection(v.Get(name))
}

// Equals returns true if the request parameter 'paramName' is equal to
// 'value'
func (v Values) Equals(paramName, value string) bool {
  return v.Get(paramName) == value
}

// Mux is the interface that wraps the Handle method.
type Mux interface {
  Handle(pattern string, handler http.Handler)
}

// AddStatic adds static content to mux.
// path is the path to the file; content is the file content.
func AddStatic(mux Mux, path, content string) {
  AddStaticBinary(mux, path, []byte(content))
}

// AddStaticBinary adds static content to mux.
// path is the path to the file; content is the file content.
func AddStaticBinary(mux Mux, path string, content []byte) {
  mux.Handle(
      path, 
      http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.ServeContent(w, r, path, kAppStart, bytes.NewReader(content))
      }))
}

// AddStaticFromFile adds static content to mux. path is the
// path to the file; localPath is the actual path of the file on the local
// filesystem.
func AddStaticFromFile(mux Mux, path, localPath string) error {
  file, err := os.Open(localPath)
  if err != nil {
    return err
  }
  defer file.Close()
  buffer := bytes.Buffer{}
  buffer.ReadFrom(file)
  AddStaticBinary(mux, path, buffer.Bytes())
  return nil
}

// Error sends the status code along with its corresponding message
func Error(w http.ResponseWriter, status int) {
  http.Error(w, fmt.Sprintf("%d %s", status, http.StatusText(status)), status)
}

func init() {
  kLog = log.New(os.Stderr, "", log.LstdFlags | log.Lmicroseconds)
  kAppStart = time.Now()
}
