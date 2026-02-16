package helpers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/test"
)

// parseRange parses a ranges header in the format "bytes=from-to" and returns
// x and y as uint64.
func parseRange(t *testing.T, str string) (from, to uint64) {
	if !strings.HasPrefix(str, "bytes=") {
		t.Fatalf("byte range %s does not start with 'bytes='", str)
	}

	str = strings.TrimPrefix(str, "bytes=")
	ranges := strings.Split(str, ",")
	if len(ranges) != 1 {
		t.Fatalf("byte range %s must have one range", str)
	}

	rng := strings.Split(ranges[0], "-")
	if len(rng) != 2 {
		t.Fatalf("byte range %s is invalid", str)
	}

	var err error
	from, err = strconv.ParseUint(rng[0], 10, 0)
	if err != nil {
		t.Fatalf("cannot parse range %s: %s", str, err.Error())
	}

	to, err = strconv.ParseUint(rng[1], 10, 0)
	if err != nil {
		t.Fatalf("cannot parse range %s: %s", str, err.Error())
	}

	return from, to
}

// combineRanges combines the multiple request ranges into a single Range header.
func combineRanges(t *testing.T, ranges []string) string {
	var str strings.Builder
	str.WriteString("bytes=")

	for i, rng := range ranges {
		from, to := parseRange(t, rng)
		str.WriteString(fmt.Sprintf("%d-%d", from, to))
		if i != len(ranges)-1 {
			str.WriteString(",")
		}
	}

	return str.String()
}

// SingleRangeTestTransform takes a test where there is no "Range" header set in the request, or checks on the
// StatusCode, Body, or Content-Range headers and verifies whether a valid response is given for the requested range.
//
// Note: HTTP Range requests can be validly responded with either the full data, or the requested partial data.
func SingleRangeTestTransform(t *testing.T, baseTest test.SugarTest, byteRange string, fullData []byte) test.SugarTest {
	modifiedRequest := baseTest.Request.Clone().Header("Range", byteRange)
	if baseTest.Requests != nil {
		t.Fatal("does not support multiple requests or responses")
	}
	modifiedResponse := baseTest.Response.Clone()

	fullSize := int64(len(fullData))
	start, end := parseRange(t, byteRange)

	rangeTest := test.SugarTest{
		Name:     baseTest.Name,
		Hint:     baseTest.Hint,
		Request:  modifiedRequest,
		Requests: nil,
		Response: test.AllOf(
			modifiedResponse,
			test.AnyOf(
				test.Expect().Status(http.StatusPartialContent).Body(fullData[start:end+1]).Headers(
					test.Header("Content-Range").Equals("bytes {{start}}-{{end}}/{{length}}", start, end, fullSize),
				),
				test.Expect().Status(http.StatusOK).Body(fullData),
			),
		),
	}

	return rangeTest
}

// MultiRangeTestTransform takes a test where there is no "Range" header set in the request, or checks on the
// StatusCode, Body, or Content-Range or Content-Type headers and verifies whether a valid response is given for the
// requested ranges.
//
// If contentType is empty it is ignored.
//
// Note: HTTP Multi Range requests can be validly responded with one of the full data, the partial data from the first
// range, or the partial data from all the requested ranges.
func MultiRangeTestTransform(t *testing.T, baseTest test.SugarTest, byteRanges []string, fullData []byte, contentType string) test.SugarTest {
	modifiedRequest := baseTest.Request.Clone().Header("Range", combineRanges(t, byteRanges))
	if baseTest.Requests != nil {
		t.Fatal("does not support multiple requests or responses")
	}
	modifiedResponse := baseTest.Response.Clone()

	fullSize := int64(len(fullData))
	type rng struct {
		start, end uint64
	}

	var multirangeBodyChecks []check.Check[string]
	var ranges []rng
	for _, r := range byteRanges {
		start, end := parseRange(t, r)
		ranges = append(ranges, rng{start: start, end: end})
		multirangeBodyChecks = append(multirangeBodyChecks,
			check.Contains("Content-Range: bytes {{start}}-{{end}}/{{length}}", ranges[0].start, ranges[0].end, fullSize),
			check.Contains(string(fullData[start:end+1])),
		)
	}

	rangeTest := test.SugarTest{
		Name:     baseTest.Name,
		Hint:     baseTest.Hint,
		Request:  modifiedRequest,
		Requests: nil,
		Response: test.AllOf(
			modifiedResponse,
			test.AnyOf(
				test.Expect().Status(http.StatusOK).Body(fullData).Header(test.Header("Content-Type", contentType)),
				test.Expect().Status(http.StatusPartialContent).Body(fullData[ranges[0].start:ranges[0].end+1]).Headers(
					test.Header("Content-Range").Equals("bytes {{start}}-{{end}}/{{length}}", ranges[0].start, ranges[0].end, fullSize),
					test.Header("Content-Type", contentType),
				),
				test.Expect().Status(http.StatusPartialContent).Body(
					check.And(
						append([]check.Check[string]{check.Contains("Content-Type: {{contentType}}", contentType)}, multirangeBodyChecks...)...,
					),
				).Headers(test.Header("Content-Type").Contains("multipart/byteranges")),
			),
		),
	}

	return rangeTest
}

// IncludeRangeTests takes a test where there is no "Range" header set in the request, or checks on the
// StatusCode, Body, or Content-Range headers and verifies whether a valid response is given for the requested ranges.
// Will test the full request, a single range request for the first passed range as well as a multi-range request for
// all the requested ranges.
//
// If contentType is empty it is ignored.
//
// If no ranges are passed, then a panic is produced.
//
// Note: HTTP Range requests can be validly responded with either the full data, or the requested partial data
// Note: HTTP Multi Range requests can be validly responded with one of the full data, the partial data from the first
// range, or the partial data from all the requested ranges
func IncludeRangeTests(t *testing.T, baseTest test.SugarTest, byteRanges []string, fullData []byte, contentType string) test.SugarTests {
	if len(byteRanges) == 0 {
		panic("byte ranges must be defined")
	}

	return includeRangeTests(t, baseTest, byteRanges, fullData, contentType)
}

// IncludeRandomRangeTests takes a test where there is no "Range" header set in the request, or checks on the
// StatusCode, Body, or Content-Range headers and verifies whether a valid response is given for the requested ranges.
// Will test the full request, a single range request for the first passed range as well as a multi-range request for
// all the requested ranges.
//
// If contentType is empty it is ignored.
//
// If no ranges are passed then some non-overlapping ranges are automatically generated for data >= 10 bytes. Smaller
// data will produce a panic to avoid undefined behavior.
//
// Note: HTTP Range requests can be validly responded with either the full data, or the requested partial data
// Note: HTTP Multi Range requests can be validly responded with one of the full data, the partial data from the first
// range, or the partial data from all the requested ranges
func IncludeRandomRangeTests(t *testing.T, baseTest test.SugarTest, fullData []byte, contentType string) test.SugarTests {
	return includeRangeTests(t, baseTest, makeRandomByteRanges(fullData), fullData, contentType)
}

func includeRangeTests(t *testing.T, baseTest test.SugarTest, byteRanges []string, fullData []byte, contentType string) test.SugarTests {
	standardBaseRequest := baseTest.Request.Clone()
	if contentType != "" {
		standardBaseRequest = standardBaseRequest.Header("Content-Type", contentType)
	}
	standardBase := test.SugarTest{
		Name:     fmt.Sprintf("%s - full request", baseTest.Name),
		Hint:     baseTest.Hint,
		Request:  standardBaseRequest,
		Requests: baseTest.Requests,
		Response: test.AllOf(
			baseTest.Response,
			test.Expect().Status(http.StatusOK).Body(fullData),
		),
		Responses: baseTest.Responses,
	}
	rangeTests := OnlyRangeTests(t, baseTest, byteRanges, fullData, contentType)
	return append(test.SugarTests{standardBase}, rangeTests...)
}

// OnlyRangeTests takes a test where there is no "Range" header set in the request, or checks on the
// StatusCode, Body, or Content-Range headers and verifies whether a valid response is given for the requested ranges.
// Will test both a single range request for the first passed range as well as a multi-range request for all the
// requested ranges.
//
// If contentType is empty it is ignored.
//
// If no ranges are passed, then a panic is produced.
//
// Note: HTTP Range requests can be validly responded with either the full data, or the requested partial data
// Note: HTTP Multi Range requests can be validly responded with one of the full data, the partial data from the first
// range, or the partial data from all the requested ranges
func OnlyRangeTests(t *testing.T, baseTest test.SugarTest, byteRanges []string, fullData []byte, contentType string) test.SugarTests {
	if len(byteRanges) == 0 {
		panic("byte ranges must be defined")
	}

	return onlyRangeTests(t, baseTest, byteRanges, fullData, contentType)
}

// OnlyRandomRangeTests takes a test where there is no "Range" header set in the request, or checks on the
// StatusCode, Body, or Content-Range headers and verifies whether a valid response is given for the requested ranges.
// Will test both a single range request for the first passed range as well as a multi-range request for all the
// requested ranges.
//
// If contentType is empty it is ignored.
//
// If no ranges are passed then some non-overlapping ranges are automatically generated for data >= 10 bytes. Smaller
// data will produce a panic to avoid undefined behavior.
//
// Note: HTTP Range requests can be validly responded with either the full data, or the requested partial data
// Note: HTTP Multi Range requests can be validly responded with one of the full data, the partial data from the first
// range, or the partial data from all the requested ranges
func OnlyRandomRangeTests(t *testing.T, baseTest test.SugarTest, fullData []byte, contentType string) test.SugarTests {
	return onlyRangeTests(t, baseTest, makeRandomByteRanges(fullData), fullData, contentType)
}

func onlyRangeTests(t *testing.T, baseTest test.SugarTest, byteRanges []string, fullData []byte, contentType string) test.SugarTests {
	singleBaseRequest := baseTest.Request.Clone()
	if contentType != "" {
		singleBaseRequest = singleBaseRequest.Header("Content-Type", contentType)
	}

	singleBase := test.SugarTest{
		Name:      fmt.Sprintf("%s - single range", baseTest.Name),
		Hint:      baseTest.Hint,
		Request:   singleBaseRequest,
		Requests:  baseTest.Requests,
		Response:  baseTest.Response,
		Responses: baseTest.Responses,
	}
	singleRange := SingleRangeTestTransform(t, singleBase, byteRanges[0], fullData)

	multiBase := test.SugarTest{
		Name:      fmt.Sprintf("%s - multi range", baseTest.Name),
		Hint:      baseTest.Hint,
		Request:   baseTest.Request,
		Requests:  baseTest.Requests,
		Response:  baseTest.Response,
		Responses: baseTest.Responses,
	}
	multiRange := MultiRangeTestTransform(t, multiBase, byteRanges, fullData, contentType)
	return test.SugarTests{singleRange, multiRange}
}

func makeRandomByteRanges(fullData []byte) []string {
	dataLen := len(fullData)
	if dataLen < 10 {
		panic("transformation not defined for data smaller than 10 bytes")
	}

	return []string{
		"bytes=7-9",
		"bytes=1-3",
	}
}
