package vault

import (
	"encoding/base64"

	"github.com/andriantp/go-braille"
)

func Lock(data string) (string, error) {
	b64 := base64.StdEncoding.EncodeToString([]byte(data))
	encoded := braille.Encode(string(b64))

	return encoded, nil
}
