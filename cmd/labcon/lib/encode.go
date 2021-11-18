package lib

import "encoding/base32"

func Base32String(p []byte) string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(p)
}
