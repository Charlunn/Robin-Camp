package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

var (
	errJSONBodyTooLarge = errors.New("request JSON body too large")
	utf8BOM             = []byte{0xEF, 0xBB, 0xBF}
)

const maxJSONBodyBytes int64 = 1 << 20 // 1 MiB, matching gin's default

func bindJSONBody(r io.Reader, dst interface{}) error {
	limited := io.LimitReader(r, maxJSONBodyBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return err
	}
	if int64(len(data)) > maxJSONBodyBytes {
		return errJSONBodyTooLarge
	}

	data = bytes.TrimPrefix(data, utf8BOM)

	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(new(struct{})); err != io.EOF {
		if err == nil {
			return errors.New("unexpected data after JSON payload")
		}
		return err
	}

	return nil
}
