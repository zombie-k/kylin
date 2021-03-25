package hash

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
)

func HmacSHA1(key string, data string) ([]byte, string) {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	digest := mac.Sum(nil)
	return digest, hex.EncodeToString(digest)
}
