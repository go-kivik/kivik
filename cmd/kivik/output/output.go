// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package output

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/spf13/pflag"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
)

// Formatter manages output formatting.
type Formatter struct {
	mu         sync.Mutex
	formats    map[string]Format
	formatOpts []string

	format    string
	output    string
	overwrite bool
}

// New returns an output formatter instance.
func New() *Formatter {
	return &Formatter{
		formats: map[string]Format{},
	}
}

// Format is the output format interface.
type Format interface {
	Output(io.Writer, io.Reader) error
}

// FormatArg is an optional interface. If implemented by a formatter, it
// may receive an argument.
type FormatArg interface {
	Arg(string) error
	Required() bool
}

// Register registers an output formatter.
func (f *Formatter) Register(name string, fmt Format) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.formats[name]; ok {
		panic(name + " already registered")
	}
	f.formats[name] = fmt
	if name != "" {
		f.formatOpts = append(f.formatOpts, formatOptions(name, fmt))
	}
}

func (f *Formatter) options() []string {
	if len(f.formats) == 0 {
		panic("no formatters regiestered")
	}
	return f.formatOpts
}

func formatOptions(name string, f Format) string {
	if argFmt, ok := f.(FormatArg); ok {
		if argFmt.Required() {
			return name + "=..."
		}
		return name + "[=...]"
	}
	return name
}

// ConfigFlags sets up the CLI flags based on the configured formatters.
func (f *Formatter) ConfigFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&f.format, "format", "f", "", "Output format. One of: "+strings.Join(f.options(), "|"))
	fs.StringVarP(&f.output, "output", "o", "", "Output file/directory.")
	fs.BoolVarP(&f.overwrite, "overwrite", "F", false, "Overwrite output file")
}

func (f *Formatter) Output(r io.Reader) error {
	fmt, err := f.formatter()
	if err != nil {
		return err
	}
	out, err := f.writer()
	if err != nil {
		return err
	}
	if c, ok := out.(io.Closer); ok {
		defer c.Close() // nolint:errcheck
	}
	return fmt.Output(out, r)
}

func (f *Formatter) formatter() (Format, error) {
	args := strings.SplitN(f.format, "=", 2) //nolint:gomnd
	name := args[0]
	if format, ok := f.formats[name]; ok {
		if fmtArg, ok := format.(FormatArg); ok {
			if fmtArg.Required() && len(args) == 1 {
				return nil, errors.Codef(errors.ErrUsage, "format %s requires an argument", name)
			}
			if len(args) > 1 {
				if err := fmtArg.Arg(args[1]); err != nil {
					return nil, errors.Code(errors.ErrUsage, err)
				}
			}
		} else if len(args) > 1 {
			return nil, errors.Codef(errors.ErrUsage, "format %s takes no arguments", name)
		}

		return format, nil
	}

	return nil, errors.Codef(errors.ErrUsage, "unrecognized output format option: %s", name)
}

func (f *Formatter) writer() (io.Writer, error) {
	switch f.output {
	case "", "-":
		return ensureNewlineEnding(os.Stdout), nil
	}
	return f.createFile(f.output)
}

func (f *Formatter) createFile(path string) (*os.File, error) {
	if f.overwrite {
		return os.Create(path)
	}
	return os.OpenFile(path, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0o666) //nolint:gomnd
}

func (f *Formatter) OK() error {
	format := `OK`
	data := struct {
		OK bool
	}{
		OK: true,
	}
	result := TemplateReader(format, data, JSONReader(map[string]bool{"ok": true}))
	return f.Output(result)
}

func (f *Formatter) UpdateResult(id, rev string) error {
	type result struct {
		OK  bool   `json:"ok"`
		ID  string `json:"id"`
		Rev string `json:"rev"`
	}

	update := result{
		OK:  true,
		ID:  id,
		Rev: rev,
	}

	format := `OK: {{ .OK }}
ID: {{ .ID }}
Rev: {{ .Rev }}`
	return f.Output(TemplateReader(format, update, JSONReader(update)))
}

func ensureNewlineEnding(w io.Writer) io.WriteCloser {
	return &addNewlineEnding{Writer: w}
}

type addNewlineEnding struct {
	io.Writer
	last byte
}

func (w *addNewlineEnding) Write(p []byte) (int, error) {
	if len(p) > 0 {
		w.last = p[len(p)-1]
	}
	return w.Writer.Write(p)
}

func (w *addNewlineEnding) Close() error {
	if w.last != '\n' {
		_, err := w.Writer.Write([]byte{'\n'})
		if err != nil {
			return err
		}
	}
	if c, ok := w.Writer.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
