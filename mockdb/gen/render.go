package main

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"
)

var tmpl *template.Template

func initTemplates(root string) {
	var err error
	tmpl, err = template.ParseGlob(root + "/*")
	if err != nil {
		panic(err)
	}
}

func RenderExpectationsGo(filename string, methods []*Method) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(file, "expectations.go.tmpl", methods)
}

func RenderClientGo(filename string, methods []*Method) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(file, "client.go.tmpl", methods)
}

func RenderMockGo(filename string, methods []*Method) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(file, "mock.go.tmpl", methods)
}

func RenderDriverMethod(m *Method) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.ExecuteTemplate(buf, "drivermethod.tmpl", m)
	return buf.String(), err
}

func RenderExpectedType(m *Method) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.ExecuteTemplate(buf, "expectedtype.tmpl", m)
	return buf.String(), err
}

func (m *Method) DriverArgs() string {
	args := make([]string, 0, len(m.Accepts)+2)
	if m.AcceptsContext {
		args = append(args, "ctx context.Context")
	}
	for i, arg := range m.Accepts {
		args = append(args, fmt.Sprintf("arg%d %s", i, typeName(arg)))
	}
	if m.AcceptsOptions {
		args = append(args, "options map[string]interface{}")
	}
	return strings.Join(args, ", ")
}

func (m *Method) ReturnArgs() string {
	args := make([]string, 0, len(m.Returns)+1)
	for _, arg := range m.Returns {
		args = append(args, arg.String())
	}
	if m.ReturnsError {
		args = append(args, "error")
	}
	if len(args) > 1 {
		return `(` + strings.Join(args, ", ") + `)`
	}
	return args[0]
}

func (m *Method) VariableDefinitions() string {
	var result []string
	for i, arg := range m.Accepts {
		result = append(result, fmt.Sprintf("\targ%d %s\n", i, typeName(arg)))
	}
	for i, ret := range m.Returns {
		name := typeName(ret)
		switch name {
		case "driver.DB": // nolint: goconst
			name = "*DB"
		case "driver.Replication": // nolint: goconst
			name = "*Replication"
		case "[]driver.Replication": // nolint: goconst
			name = "[]*Replication"
		}
		result = append(result, fmt.Sprintf("\tret%d %s\n", i, name))
	}
	return strings.Join(result, "")
}

func (m *Method) inputVars() []string {
	args := make([]string, 0, len(m.Accepts)+1)
	for i := range m.Accepts {
		args = append(args, fmt.Sprintf("arg%d", i))
	}
	if m.AcceptsOptions {
		args = append(args, "options")
	}
	return args
}

func (m *Method) ExpectedVariables() string {
	args := []string{}
	if m.DBMethod {
		args = append(args, "db")
	}
	args = append(args, m.inputVars()...)
	return alignVars(0, args)
}

func (m *Method) InputVariables() string {
	var result, common []string
	if m.DBMethod {
		common = append(common, "\t\t\tdb: db.DB,\n")
	}
	for i := range m.Accepts {
		result = append(result, fmt.Sprintf("\t\targ%d: arg%d,\n", i, i))
	}
	if m.AcceptsOptions {
		common = append(common, "\t\t\toptions: options,\n")
	}
	if len(common) > 0 {
		result = append(result, fmt.Sprintf("\t\tcommonExpectation: commonExpectation{\n%s\t\t},\n",
			strings.Join(common, "")))
	}
	return strings.Join(result, "")
}

func (m *Method) Variables(indent int) string {
	args := m.inputVars()
	for i := range m.Returns {
		args = append(args, fmt.Sprintf("ret%d", i))
	}
	return alignVars(indent, args)
}

func alignVars(indent int, args []string) string {
	var maxLen int
	for _, arg := range args {
		if l := len(arg); l > maxLen {
			maxLen = l
		}
	}
	final := make([]string, len(args))
	for i, arg := range args {
		final[i] = fmt.Sprintf("%s%*s %s,", strings.Repeat("\t", indent), -(maxLen + 1), arg+":", arg)
	}
	return strings.Join(final, "\n")
}

func (m *Method) ZeroReturns() string {
	args := make([]string, 0, len(m.Returns))
	for _, arg := range m.Returns {
		args = append(args, zeroValue(arg))
	}
	args = append(args, "err")
	return strings.Join(args, ", ")
}

func zeroValue(t reflect.Type) string {
	z := fmt.Sprintf("%#v", reflect.Zero(t).Interface())
	if strings.HasSuffix(z, "(nil)") {
		return "nil"
	}
	switch z {
	case "<nil>":
		return "nil"
	}
	return z
}

func (m *Method) ExpectedReturns() string {
	args := make([]string, 0, len(m.Returns))
	for i, arg := range m.Returns {
		switch arg.String() {
		case "driver.Rows":
			args = append(args, fmt.Sprintf("&driverRows{Context: ctx, Rows: coalesceRows(expected.ret%d)}", i))
		case "driver.Changes":
			args = append(args, fmt.Sprintf("&driverChanges{Context: ctx, Changes: coalesceChanges(expected.ret%d)}", i))
		case "driver.DB":
			args = append(args, fmt.Sprintf("&driverDB{DB: expected.ret%d}", i))
		case "driver.DBUpdates":
			args = append(args, fmt.Sprintf("&driverDBUpdates{Context:ctx, Updates: coalesceDBUpdates(expected.ret%d)}", i))
		case "driver.Replication":
			args = append(args, fmt.Sprintf("&driverReplication{Replication: expected.ret%d}", i))
		case "[]driver.Replication":
			args = append(args, fmt.Sprintf("driverReplications(expected.ret%d)", i))
		default:
			args = append(args, fmt.Sprintf("expected.ret%d", i))
		}
	}
	if m.AcceptsContext {
		args = append(args, "expected.wait(ctx)")
	} else {
		args = append(args, "expected.err")
	}
	return strings.Join(args, ", ")
}

func (m *Method) ReturnTypes() string {
	args := make([]string, len(m.Returns))
	for i, ret := range m.Returns {
		name := typeName(ret)
		switch name {
		case "driver.DB":
			name = "*DB"
		case "driver.Replication":
			name = "*Replication"
		case "[]driver.Replication":
			name = "[]*Replication"
		}
		args[i] = fmt.Sprintf("ret%d %s", i, name)
	}
	return strings.Join(args, ", ")
}

func typeName(t reflect.Type) string {
	name := t.String()
	switch name {
	case "interface {}":
		return "interface{}"
	case "driver.Rows":
		return "*Rows"
	case "driver.Changes":
		return "*Changes"
	case "driver.DBUpdates":
		return "*Updates"
	}
	return name
}

func (m *Method) SetExpectations() string {
	var args []string
	if m.DBMethod {
		args = append(args, "commonExpectation: commonExpectation{db: db},\n")
	}
	if m.Name == "DB" {
		args = append(args, "ret0: &DB{},\n")
	}
	for i, ret := range m.Returns {
		var zero string
		switch ret.String() {
		case "*kivik.Rows":
			zero = "&Rows{}"
		case "*kivik.QueryPlan":
			zero = "&driver.QueryPlan{}"
		case "*kivik.PurgeResult":
			zero = "&driver.PurgeResult{}"
		case "*kivik.DBUpdates":
			zero = "&Updates{}"
		}
		if zero != "" {
			args = append(args, fmt.Sprintf("ret%d: %s,\n", i, zero))
		}
	}
	return strings.Join(args, "")
}

func (m *Method) MetExpectations() string {
	if len(m.Accepts) == 0 {
		return ""
	}
	args := make([]string, 0, len(m.Accepts)+1)
	args = append(args, fmt.Sprintf("\texp := ex.(*Expected%s)", m.Name))
	var check string
	for i, arg := range m.Accepts {
		switch arg.String() {
		case "string":
			check = `exp.arg%[1]d != "" && exp.arg%[1]d != e.arg%[1]d`
		case "int":
			check = "exp.arg%[1]d != 0 && exp.arg%[1]d != e.arg%[1]d"
		case "interface {}":
			check = "exp.arg%[1]d != nil && !jsonMeets(exp.arg%[1]d, e.arg%[1]d)"
		default:
			check = "exp.arg%[1]d != nil && !reflect.DeepEqual(exp.arg%[1]d, e.arg%[1]d)"
		}
		args = append(args, fmt.Sprintf("if "+check+" {\n\t\treturn false\n\t}", i))
	}
	return strings.Join(args, "\n")
}

func (m *Method) MethodArgs() string {
	var args, vars, str, def, mid []string
	prefix := ""
	if m.DBMethod {
		prefix = "DB(%s)."
		args = append(args, "e.dbo().name")
	}
	if m.AcceptsContext {
		vars = append(vars, "ctx")
	}
	var lines []string
	for i, acc := range m.Accepts {
		str = append(str, fmt.Sprintf("arg%d", i))
		def = append(def, `"?"`)
		vars = append(vars, "%s")
		switch acc.String() {
		case "string":
			mid = append(mid, fmt.Sprintf(`	if e.arg%[1]d != "" { arg%[1]d = fmt.Sprintf("%%q", e.arg%[1]d)}`, i))
		case "int":
			mid = append(mid, fmt.Sprintf(`	if e.arg%[1]d != 0 { arg%[1]d = fmt.Sprintf("%%q", e.arg%[1]d)}`, i))
		default:
			mid = append(mid, fmt.Sprintf(`	if e.arg%[1]d != nil { arg%[1]d = fmt.Sprintf("%%v", e.arg%[1]d) }`, i))
		}
	}
	if m.AcceptsOptions {
		str = append(str, "options")
		def = append(def, `defaultOptionPlaceholder`)
		vars = append(vars, "%s")
		mid = append(mid, `	if e.options != nil { options = fmt.Sprintf("%v", e.options) }`)
	}
	if len(str) > 0 {
		lines = append(lines, fmt.Sprintf("\t%s := %s", strings.Join(str, ", "), strings.Join(def, ", ")))
	}
	lines = append(lines, mid...)
	lines = append(lines, fmt.Sprintf("\treturn fmt.Sprintf(\"%s%s(%s)\", %s)", prefix, m.Name, strings.Join(vars, ", "), strings.Join(append(args, str...), ", ")))
	return strings.Join(lines, "\n")
}

// CallbackType returns the type definition for a callback for this method.
func (m *Method) CallbackTypes() string {
	inputs := make([]string, 0, len(m.Accepts)+2)
	if m.AcceptsContext {
		inputs = append(inputs, "context.Context")
	}
	for _, arg := range m.Accepts {
		inputs = append(inputs, typeName(arg))
	}
	if m.AcceptsOptions {
		inputs = append(inputs, "map[string]interface{}")
	}
	return strings.Join(inputs, ", ")
}

// CallbackArgs returns the list of arguments to be passed to the callback
func (m *Method) CallbackArgs() string {
	args := make([]string, 0, len(m.Accepts)+2)
	if m.AcceptsContext {
		args = append(args, "ctx")
	}
	for i := range m.Accepts {
		args = append(args, fmt.Sprintf("arg%d", i))
	}
	if m.AcceptsOptions {
		args = append(args, "options")
	}
	return strings.Join(args, ", ")
}

func (m *Method) CallbackReturns() string {
	args := make([]string, 0, len(m.Returns)+1)
	for _, ret := range m.Returns {
		args = append(args, ret.String())
	}
	if m.ReturnsError {
		args = append(args, "error")
	}
	if len(args) > 1 {
		return "(" + strings.Join(args, ", ") + ")"
	}
	return strings.Join(args, ", ")
}
