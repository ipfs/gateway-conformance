package check

import (
	"fmt"

	"github.com/ipfs/gateway-conformance/tooling/ipns"
)

var _ Check[[]byte] = &CheckIsIPNSKey{}

type CheckIsIPNSKey struct {
	shouldBeValid bool
	expectedValue string
}

func IsIPNSKey() *CheckIsIPNSKey {
	return &CheckIsIPNSKey{
		shouldBeValid: true,
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

func (c *CheckIsIPNSKey) Check(ipnsKey []byte) CheckOutput {
	record, err := ipns.UnmarshalIpnsRecord(ipnsKey)

	if err != nil {
		if c.shouldBeValid {
			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("IPNS key '%s' is not valid: %v", ipnsKey, err),
			}
		} else {
			panic("not implemented")
		}
	}

	if c.expectedValue != "" {
		if record.Value() != c.expectedValue {
			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("IPNS key '%s' points to '%s', but expected value is '%s'", ipnsKey, record.Value(), c.expectedValue),
			}
		}
	}

	return CheckOutput{
		Success: true,
	}
}