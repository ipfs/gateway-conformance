package ipns

import (
	"fmt"

	"github.com/ipfs/gateway-conformance/tooling/check"
)

var _ check.Check[[]byte] = &CheckIsIPNSKey{}

type CheckIsIPNSKey struct {
	shouldBeValid bool
	expectedValue string
	pubKey         string
}

func IsIPNSKey(keyId string) *CheckIsIPNSKey {
	return &CheckIsIPNSKey{
		shouldBeValid: true,
		pubKey:         keyId,
		expectedValue: "",
	}
}

func (c *CheckIsIPNSKey) IsValid() *CheckIsIPNSKey {
	c.shouldBeValid = true
	return c
}

func (c *CheckIsIPNSKey) PointsTo(value string, rest ...any) *CheckIsIPNSKey {
	c.expectedValue = fmt.Sprintf(value, rest...)
	return c
}

func (c *CheckIsIPNSKey) Check(recordPayload []byte) check.CheckOutput {
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