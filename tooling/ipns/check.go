package ipns

import (
	"fmt"

	"github.com/ipfs/gateway-conformance/tooling/check"
)

var _ check.Check[[]byte] = &CheckIsIPNSRecord{}

type CheckIsIPNSRecord struct {
	shouldBeValid bool
	expectedValue string
	pubKey        string
}

func IsIPNSRecord(keyId string) *CheckIsIPNSRecord {
	return &CheckIsIPNSRecord{
		shouldBeValid: true,
		pubKey:        keyId,
		expectedValue: "",
	}
}

func (c *CheckIsIPNSRecord) IsValid() *CheckIsIPNSRecord {
	c.shouldBeValid = true
	return c
}

func (c *CheckIsIPNSRecord) PointsTo(value string, rest ...any) *CheckIsIPNSRecord {
	c.expectedValue = fmt.Sprintf(value, rest...)
	return c
}

func (c *CheckIsIPNSRecord) Check(recordPayload []byte) check.CheckOutput {
	record, err := UnmarshalIpnsRecord(recordPayload, c.pubKey)

	if err != nil {
		if c.shouldBeValid {
			return check.CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("IPNS key '%s' is not valid: %v", recordPayload, err),
			}
		} else {
			panic("not implemented")
		}
	}

	if c.expectedValue != "" {
		if record.Value() != c.expectedValue {
			return check.CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("IPNS key '%s' points to '%s', but expected value is '%s'", recordPayload, record.Value(), c.expectedValue),
			}
		}
	}

	return check.CheckOutput{
		Success: true,
	}
}
