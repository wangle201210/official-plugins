package watermark

import (
	"crypto/md5"
	"encoding/hex"
)

const (
	offset64 = uint64(14695981039346656037)
	prime64  = uint64(1099511628211)
)

// MD5 returns the md5 string according to the given string
func MD5(str string) string {

	return MD5Bytes([]byte(str))
}

func MD5Bytes(buf []byte) string {
	h := md5.New()
	h.Write(buf)
	return hex.EncodeToString(h.Sum(nil))
}

func Hash(str string) uint64 {

	offset := 0
	n := (len(str) / 8) * 8
	hash := offset64

	for offset = 0; offset < n; offset += 8 {
		hash = (hash ^ uint64(str[offset])) * prime64
		hash = (hash ^ uint64(str[offset+1])) * prime64
		hash = (hash ^ uint64(str[offset+2])) * prime64
		hash = (hash ^ uint64(str[offset+3])) * prime64
		hash = (hash ^ uint64(str[offset+4])) * prime64
		hash = (hash ^ uint64(str[offset+5])) * prime64
		hash = (hash ^ uint64(str[offset+6])) * prime64
		hash = (hash ^ uint64(str[offset+7])) * prime64
	}

	for _, c := range str[offset:] {
		hash = (hash ^ uint64(c)) * prime64
	}

	return hash
}

func SimpleHash(str string) uint64 {

	var sum uint64
	buf := []byte(str)
	for _, b := range buf {
		sum += uint64(b)
	}

	return sum
}
