package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"testing"
	"text/template"
)

type ReportInput struct {
	Req  *http.Request
	Res  *http.Response
	Err  error
	Test SugarTest
}

const TEMPLATE = `
Name: {{.Test.Name}}
Hint: {{.Test.Hint}}

Error: {{.Err}}

Expected Request:
{{.Test.Request | json}}

Actual Request:
{{.Req | dump}}

Expected Response:
{{.Test.Response | json}}

Actual Response:
{{.Res | dump}}
`

func report(t *testing.T, test SugarTest, req *http.Request, res *http.Response, err error) {
	t.Helper()

	input := ReportInput{
		Req:  req,
		Res:  res,
		Err:  err,
		Test: test,
	}

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"json": func(v any) string {
			j, _ := json.MarshalIndent(v, "", "  ")
			return string(j)
		},
		"dump": func(v any) string {
			if v == nil {
				return "nil"
			}

			var b []byte
			var err error
			switch v := v.(type) {
			case *http.Request:
				if v == nil {
					return "nil" // golang does not catch the nil case above
				}

				b, err = httputil.DumpRequestOut(v, true)
			case *http.Response:
				if v == nil {
					return "nil" // golang does not catch the nil case above
				}
				// TODO: we have to disable the body dump because
				// it triggers an error:
				// "http: ContentLength=6 with Body length 0"
				b, err = httputil.DumpResponse(v, false)
			default:
				return fmt.Sprintf("error: unknown type %T", v)
			}

			if err != nil {
				return fmt.Sprintf("error: failed to dump %T: %v", v, err)
			}

			return string(b)
		},
	}).Parse(TEMPLATE)

	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, input)
	if err != nil {
		panic(fmt.Errorf("failed to execute template: %#v %w", input, err))
	}

	if input.Err != nil {
		t.Fatal(buf.String())
	} else {
		t.Log(buf.String())
	}
}
