package common

import (
	"encoding/json"
	"errors"
)

// Metadata enforces the types allowed as values in Layer metadata
type Metadata struct {
	data map[string]interface{}
}

// Set adds data to your Metadata so long as the value is a string or Metadata.
func (m *Metadata) Set(key string, value interface{}) error {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}

	switch value.(type) {
	case string:
		m.data[key] = value
	case Metadata:
		m.data[key] = value
	default:
		return errors.New("value must be string or Metadata")
	}
	return nil
}

// Get retrieves the value in the map and returns nil if the key is not found.
func (m *Metadata) Get(key string) interface{} {
	if m.data == nil {
		return nil
	}
	return m.data[key]
}

func (m Metadata) ToMap() map[string]string {
	strMap := make(map[string]string)
	if m.data == nil {
		return strMap
	}

	for key, value := range m.data {
		switch value.(type) {
		case string:
			strMap[key] = value.(string)
		}
	}
	return strMap
}

func (m Metadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.data)
}

func (m *Metadata) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &m.data)
}
