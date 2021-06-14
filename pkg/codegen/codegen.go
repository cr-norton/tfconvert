package codegen

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/cr-norton/tfconvert/pkg/templates"
	"github.com/cr-norton/tfconvert/pkg/types"
	"github.com/pkg/errors"
)

func Generate(stack Stack, options types.Options, functions map[string]interface{}) (map[string]string, error) {
	tmpl, err := parseTemplate(functions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse templates")
	}

	srcs, err := listTemplates()
	if err != nil {
		return nil, err
	}

	tfout := map[string]string{}
	for _, src := range srcs {
		out, err := executeTemplate(tmpl, src, stack)
		if err != nil {
			return nil, errors.Wrap(err, "unable to execute template")
		}
		name := outName(src)
		tfout[name] = string(out)
	}
	return tfout, nil
}

func parseTemplate(funcs map[string]interface{}) (*template.Template, error) {
	funcs = mergeTemplateFunctions(funcs)
	tmpl := template.New("tfconvert").Funcs(funcs)
	tmpl, err := templates.Parse(tmpl)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse templates")
	}
	return tmpl, nil
}

func listTemplates() ([]string, error) {
	dir := fmt.Sprintf("%s/src/github.com/cr-norton/tfconvert/pkg/templates", os.Getenv("GOPATH"))
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list templates")
	}
	paths := []string{}
	for _, info := range infos {
		if strings.HasSuffix(info.Name(), ".tmpl") {
			paths = append(paths, info.Name())
		}
	}
	return paths, nil
}

func executeTemplate(t *template.Template, name string, params interface{}) ([]byte, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	if err := t.ExecuteTemplate(w, name, params); err != nil {
		return nil, err
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func outName(fname string) string {
	return strings.Replace(fname, ".tmpl", ".tf", 1)
}
