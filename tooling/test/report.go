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

// If the response is invalid (the content length > Body length for example),
// go's DumpResponse will panic. This function recover's from that panic and
// dumps the response again without the body.
func safeDumpResponse(res *http.Response) (b []byte, err error) {
	if res == nil {
		return []byte("nil"), nil
	}

	// Attempt to dump the response with the body included
	defer func() {
		if r := recover(); r != nil {
			// If a panic occurred, dump the response again without the body
			b, err = httputil.DumpResponse(res, false)
		}
	}()

	b, err = httputil.DumpResponse(res, true)

	return b, err
}

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
				b, err = safeDumpResponse(v)
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
