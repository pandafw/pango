package gingzip

import (
	"testing"
)

func TestGzipProxiedFlag(t *testing.T) {
	fs := []ProxiedFlag{
		ProxiedOff,
		ProxiedAny,
		ProxiedAuth,
		ProxiedExpired,
		ProxiedNoCache,
		ProxiedNoStore,
		ProxiedPrivate,
		ProxiedNoLastModified,
		ProxiedNoETag,
	}

	for i, f := range fs {
		e := 0
		if i > 0 {
			e = 1 << (i - 1)
		}
		a := int(f)
		if e != a {
			t.Errorf("%v = %v, want %v", f, a, e)
		}
	}
}
