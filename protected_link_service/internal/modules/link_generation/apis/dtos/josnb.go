package apiDtos

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JSONB map[string]interface{}

func (j JSONB) UnmarshalJSON(b []byte) any {
	panic("unimplemented")
}

// Value converts JSONB to a database-compatible format.
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan converts database JSONB to the JSONB struct.
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONB")
	}
	return json.Unmarshal(bytes, j)
}
func StructToJSONB(input interface{}, v any) error {
	// Marshal struct to JSON bytes
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return err
	}

	// Unmarshal JSON bytes into map[string]interface{}

	err = json.Unmarshal(jsonBytes, v)
	if err != nil {
		return err
	}
	return nil

}
