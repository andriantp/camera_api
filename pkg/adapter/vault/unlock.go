package vault

import (
	"encoding/base64"
	"fmt"

	"github.com/andriantp/go-braille"
)

func Unlock(data string) ([]byte, error) {
	decoded := braille.Decode(data)

	result, err := base64.StdEncoding.DecodeString(decoded)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}

	return result, nil
}
