package base20

import "strings"

var base = []byte("23456789CFGHJMPQRVWX")

func Encode(i int64) string {
	var b strings.Builder
	for i > 0 {
		b.WriteByte(base[i%20])
		i /= 20
	}
	return b.String()
}

func Decode(s string) int64 {
	var i, mul int64 = 0, 1
	for _, r := range []byte(s) {
		i += idx(r) * mul
		mul *= 20
	}
	return i
}

func idx(r byte) int64 {
	for i, s := range base {
		if s == r {
			return int64(i)
		}
	}
	return -1
}
