package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
}

func TestCreateJSONString(t *testing.T) {
	// Example usage of ToJSONString.
	props := DomainProperties{
		DomainName: "example.com",
		StringProperties: map[string]string{
			"owner": "John Doe",
			"email": "john@example.com",
		},
		IntProperties: map[string]int64{
			"creationDate":   1620784027,
			"expirationDate": 1652320027,
		},
		FloatProperties: map[string]float64{
			"rating": 4.5,
		},
	}

	jsonString, err := props.ToJSONString()
	if err != nil {
		assert.Failf(t, "Unable to marshal JSON, err: %s", err.Error())
		return
	}

	expectedOutput := "{\"creationDate\":1620784027,\"domainName\":\"example.com\"," +
		"\"email\":\"john@example.com\",\"expirationDate\":1652320027,\"owner\":\"John Doe\",\"rating\":4.5}"
	assert.Equal(t, jsonString, expectedOutput)
}
