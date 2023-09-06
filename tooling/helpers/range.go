package helpers

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/test"
)

// ByteRange describes an HTTP range request and the data it corresponds to. "From" and "To" mostly
// follow [HTTP Byte Range] Request semantics:
//
//   - From >= 0 and To = nil: Get the file (From, Length)
//   - From >= 0 and To >= 0: Get the range (From, To)
//   - From >= 0 and To <0: Get the range (From, Length - To)
//
// [HTTP Byte Range]: https://httpwg.org/specs/rfc9110.html#rfc.section.14.1.2
type ByteRange struct {
	From       uint64
	To         *int64
	RangeBytes []byte
}

func SimpleByteRange(from, to uint64, data []byte) ByteRange {
	toInt := int64(to)
	return ByteRange{
		From:       from,
		To:         &toInt,
		RangeBytes: data,
	}
}

func (b ByteRange) GetRangeString(t *testing.T) string {
	strWithoutPrefix := b.getRangeStringWithoutPrefix(t)
	return fmt.Sprintf("bytes=%s", strWithoutPrefix)
}

func (b ByteRange) getRangeStringWithoutPrefix(t *testing.T) string {
	if b.To == nil {
		return fmt.Sprintf("%d-", b.From)
	}

	to := *b.To
	if to >= 0 {
		return fmt.Sprintf("%d-%d", b.From, to)
	}

	if to < 0 && b.From != 0 {
		t.Fatalf("for a suffix request the From field must be 0")
	}
	return fmt.Sprintf("%d", to)
}

func (b ByteRange) getRange(t *testing.T, totalSize int64) (uint64, uint64) {
	if totalSize < 0 {
		t.Fatalf("total size must be greater than 0")
	}

	if b.To == nil {
		return b.From, uint64(totalSize)
	}

	to := *b.To
	if to >= 0 {
		return b.From, uint64(to)
	}

	if to < 0 && b.From != 0 {
		t.Fatalf("for a suffix request the From field must be 0")
	}

	start := int64(totalSize) + to
	if start < 0 {
		t.Fatalf("suffix request must not start before the start of the file")
	}

	return uint64(start), uint64(totalSize)
}

type ByteRanges []ByteRange

func (b ByteRanges) GetRangeString(t *testing.T) string {
	var rangeStrs []string
	for _, r := range b {
		rangeStrs = append(rangeStrs, r.getRangeStringWithoutPrefix(t))
	}
	return fmt.Sprintf("bytes=%s", strings.Join(rangeStrs, ","))
}

// SingleRangeTestTransform takes a test where there is no "Range" header set in the request, or checks on the
// StatusCode, Body, or Content-Range headers and verifies whether a valid response is given for the requested range.
//
// Note: HTTP Range requests can be validly responded with either the full data, or the requested partial data
func SingleRangeTestTransform(t *testing.T, baseTest test.SugarTest, brange ByteRange, fullData []byte) test.SugarTest {
	modifiedRequest := baseTest.Request.Clone().Header("Range", brange.GetRangeString(t))
	if baseTest.Requests != nil {
		t.Fatal("does not support multiple requests or responses")
	}

	fullSize := int64(len(fullData))
	start, end := brange.getRange(t, fullSize)

	rangeTest := test.SugarTest{
		Name:     baseTest.Name,
		Hint:     baseTest.Hint,
		Request:  modifiedRequest,
		Requests: nil,
		Response: test.AllOf(
			baseTest.Response,
			test.AnyOf(
				test.Expect().Status(http.StatusPartialContent).Body(brange.RangeBytes).Headers(
					test.Header("Content-Range").Equals("bytes {{start}}-{{end}}/{{length}}", start, end, fullSize),
				),
				test.Expect().Status(http.StatusOK).Body(fullData),
			),
		),
		Responses: baseTest.Responses,
	}

	return rangeTest
}

// MultiRangeTestTransform takes a test where there is no "Range" header set in the request, or checks on the
// StatusCode, Body, or Content-Range headers and verifies whether a valid response is given for the requested ranges.
//
// Note: HTTP Multi Range requests can be validly responded with one of the full data, the partial data from the first
// range, or the partial data from all the requested ranges
func MultiRangeTestTransform(t *testing.T, testWithoutRangeRequestHeader test.SugarTest, branges ByteRanges, fullData []byte) test.SugarTest {
	modifiedRequest := testWithoutRangeRequestHeader.Request.Clone().Header("Range", branges.GetRangeString(t))
	if testWithoutRangeRequestHeader.Requests != nil {
		t.Fatal("does not support multiple requests or responses")
	}

	fullSize := int64(len(fullData))
	type rng struct {
		start, end uint64
	}

	var multirangeBodyChecks []check.Check[string]
	var ranges []rng
	for _, r := range branges {
		start, end := r.getRange(t, fullSize)
		ranges = append(ranges, rng{start: start, end: end})
		multirangeBodyChecks = append(multirangeBodyChecks,
			check.Contains("Content-Range: bytes {{start}}-{{end}}/{{length}}", ranges[0].start, ranges[0].end, fullSize),
			check.Contains(string(r.RangeBytes)),
		)
	}

	rangeTest := test.SugarTest{
		Name:     testWithoutRangeRequestHeader.Name,
		Hint:     testWithoutRangeRequestHeader.Hint,
		Request:  modifiedRequest,
		Requests: nil,
		Response: test.AllOf(
			testWithoutRangeRequestHeader.Response,
			test.AnyOf(
				test.Expect().Status(http.StatusOK).Body(fullData),
				test.Expect().Status(http.StatusPartialContent).Body(branges[0].RangeBytes).Headers(
					test.Header("Content-Range").Equals("bytes {{start}}-{{end}}/{{length}}", ranges[0].start, ranges[0].end, fullSize),
				),
				test.Expect().Status(http.StatusPartialContent).Body(
					check.And(
						append([]check.Check[string]{check.Contains("Content-Type: application/vnd.ipld.raw")}, multirangeBodyChecks...)...,
					),
				).Headers(test.Header("Content-Type").Contains("multipart/byteranges")),
			),
		),
		Responses: testWithoutRangeRequestHeader.Responses,
	}

	return rangeTest
}
