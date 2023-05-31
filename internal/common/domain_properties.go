package common

import (
	"encoding/json"
)

type DomainProperties struct {
	DomainName       string
	StringProperties map[string]string
	IntProperties    map[string]int64
	FloatProperties  map[string]float64
	BoolProperties   map[string]bool
}

func NewDomainProperties() DomainProperties {
	return DomainProperties{
		DomainName:       "",
		StringProperties: make(map[string]string),
		IntProperties:    make(map[string]int64),
		FloatProperties:  make(map[string]float64),
		BoolProperties:   make(map[string]bool),
	}
}

func (d *DomainProperties) UnmarshalJSON(data []byte) error {
	// Decode the raw JSON data into a map
	var rawData map[string]interface{}
	err := json.Unmarshal(data, &rawData)
	if err != nil {
		return err
	}

	// Initialize maps
	if d.StringProperties == nil {
		d.StringProperties = make(map[string]string)
	}

	if d.IntProperties == nil {
		d.IntProperties = make(map[string]int64)
	}

	if d.FloatProperties == nil {
		d.FloatProperties = make(map[string]float64)
	}

	if d.BoolProperties == nil {
		d.BoolProperties = make(map[string]bool)
	}
	d.DomainName = rawData["domainName"].(string)

	for key, value := range rawData {
		if key == "domainName" {
			continue
		}

		switch value := value.(type) {
		case string:
			d.StringProperties[key] = value
		case float64:
			if value == float64(int64(value)) {
				d.IntProperties[key] = int64(value)
			} else {
				d.FloatProperties[key] = value
			}
		case bool:
			d.BoolProperties[key] = value
		default:
			// Ignore unknown value types
		}
	}

	return nil
}

func (d *DomainProperties) ToJSONString() (string, error) {
	// Create a map to hold the domain properties.
	properties := make(map[string]interface{})
	properties["domainName"] = d.DomainName

	for k, v := range d.StringProperties {
		properties[k] = v
	}

	for k, v := range d.IntProperties {
		properties[k] = v
	}

	for k, v := range d.FloatProperties {
		properties[k] = v
	}

	for k, v := range d.BoolProperties {
		properties[k] = v
	}

	// Marshal the map to JSON.
	jsonData, err := json.Marshal(properties)
	if err != nil {
		return "", err
	}

	// Return the JSON data as a string.
	return string(jsonData), nil
}
