package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"testing"
	"text/template"
)


type ReportInput struct {
	Req  *http.Request
	Res  *http.Response
	Err  error
	Test CTest
}

const TEMPLATE = `
Name: {{.Test.Name}}
Hint: {{.Test.Hint}}

Error: {{.Err}}

Request:
{{.Test.Request | json}}

Expected Response:
{{.Test.Response | json}}

Actual Request:
{{.Req | dump}}

Actual Response:
{{.Res | dump}}
`

func report(t *testing.T, test CTest, req *http.Request, res *http.Response, err error) {
	input := ReportInput{
		Req:  req,
		Res:  res,
		Err:  err,
		Test: test,
	}

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"json": func(v interface{}) string {
			j, _ := json.MarshalIndent(v, "", "  ")
			return string(j)
		},
		"dump": func(v interface{}) string {
			if v == nil {
				return "nil"
			}
			
			var b []byte
			var err error
			switch v := v.(type) {
			case *http.Request:
				b, err = httputil.DumpRequestOut(v, true)
			case *http.Response:
				b, err = httputil.DumpResponse(v, true)
			default:
				panic("unknown type")
			}
			if err != nil {
				panic(err)
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
		panic(err)
	}

	if input.Err != nil {
		t.Fatal(buf.String())
	} else {
		t.Log(buf.String())
	}
}
