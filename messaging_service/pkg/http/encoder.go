
package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
)

type Encoder interface {
	Encode(v interface{}) (io.Reader, ContentType, error)
}

type Decoder interface {
	Decode(r io.Reader, out interface{}) error
}

type JSONCodec struct{}

func (j *JSONCodec) Encode(v interface{}) (io.Reader, ContentType, error) {
	if v == nil {
		return nil, "", nil
	}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(v)
	return buf, ContentTypeJSON, err
}

func (j *JSONCodec) Decode(r io.Reader, out interface{}) error {
	return json.NewDecoder(r).Decode(out)
}
