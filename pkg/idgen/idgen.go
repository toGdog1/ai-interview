package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

var counter uint64

// New generates a unique, sortable identifier with an optional prefix.
func New(prefix string) string {
	n := atomic.AddUint64(&counter, 1)
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%s_%d_%d_%s", prefix, time.Now().UnixNano(), n, hex.EncodeToString(b[:]))
}
