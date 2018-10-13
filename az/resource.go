package az

import (
	"strings"
)

const (
	keySub  = "subscriptions"
	keyRG   = "resourceGroups"
	keyProv = "providers"
)

// RID is an Azure resource ID. A valid ID must begin with a '/' and contain an
// even number of path elements, which are interpreted as key/value pairs.
type RID string

// Subscription returns the resource subscription ID.
func (r RID) Subscription() string { return r.Get(keySub) }

// ResourceGroup returns the resource group name.
func (r RID) ResourceGroup() string { return r.Get(keyRG) }

// Provider returns the resource provider name.
func (r RID) Provider() string { return r.Get(keyProv) }

// Get returns the value for the specified key.
func (r RID) Get(key string) string {
	for p := string(r); p != ""; {
		i, j, k := nextPairIdx(p)
		if keyEq(p[i:j], key) {
			return p[j+1 : k]
		}
		p = p[k:]
	}
	return ""
}

// Type returns the last key in r or "<type>[/<subtype>[/...]]" if r contains a
// "/providers/" element.
func (r RID) Type() string {
	typ, _, p := nextPair(string(r))
	for p != "" {
		isProv := keyEq(typ, keyProv)
		if typ, _, p = nextPair(p); isProv && p != "" {
			var buf [128]byte
			b := append(buf[:0], typ...)
			for {
				typ, _, p = nextPair(p)
				if b = append(append(b, '/'), typ...); p == "" {
					return string(b)
				}
			}
		}
	}
	return typ
}

// Name returns the last value in r.
func (r RID) Name() string {
	return string(r[strings.LastIndexByte(string(r), '/')+1:])
}

// Norm returns a normalized representation of r.
func (r RID) Norm() RID {
	b := lazyBuilder{src: string(r)}

	// Subscription
	k, v, p := nextPair(string(r))
	if !keyEq(k, keySub) {
		r.invalid()
	}
	b.append("/" + keySub + "/").append(strings.ToLower(v))
	if p == "" {
		return RID(b.string())
	}

	// Resource group
	var haveRG bool
	if k, v, p = nextPair(p); keyEq(k, keyRG) {
		b.append("/" + keyRG + "/").append(strings.ToLower(v))
		if p == "" {
			return RID(b.string())
		}
		k, v, p = nextPair(p)
		haveRG = true
	}

	// Provider
	if !keyEq(k, keyProv) || haveRG == (p == "") {
		r.invalid()
	}
	return RID(b.append("/" + keyProv + "/").append(v).append(p).string())
}

// invalid panics to indicate an invalid resource ID format.
func (r RID) invalid() {
	panic("az: invalid resource id: " + string(r))
}

// nextIdx returns the indices of the next path component in p, starting at off.
// It panics if a non-empty p[off:] does not begin with a '/'.
func nextIdx(p string, off int) (i, j int) {
	if off < len(p) {
		if p[off] != '/' {
			panic("az: invalid resource path: " + p[off:])
		}
		i = off + 1
		if j = strings.IndexByte(p[i:], '/'); j >= 0 {
			j += i
		} else {
			j = len(p)
		}
	} else {
		i, j = off, off
	}
	return
}

// nextPairIdx returns the indices the next pair of path components in p. Index
// j is the middle separator. It panics if a non-empty p does not begin with a
// '/' or contain at least two path components.
func nextPairIdx(p string) (i, j, k int) {
	if i, j = nextIdx(p, 0); i > 0 {
		if _, k = nextIdx(p, j); k == j {
			panic("az: missing component value in resource path: " + p)
		}
	}
	return
}

// nextPair returns the next pair of path components in p as strings.
func nextPair(p string) (key, val, tail string) {
	if i, j, k := nextPairIdx(p); j < k {
		key, val, tail = p[i:j], p[j+1:k], p[k:]
	}
	return
}

// keyEq does case-insensitive comparison of ASCII strings.
func keyEq(s, t string) bool {
	return len(s) == len(t) && (s == t || strings.EqualFold(s, t))
}

// lazyBuilder is a string builder that avoids allocation if the new string is
// the same as the source.
type lazyBuilder struct {
	src string
	i   int
	b   strings.Builder
}

func (b *lazyBuilder) append(s string) *lazyBuilder {
	if b.b.Len() == 0 {
		if j := b.i + len(s); j <= len(b.src) && b.src[b.i:j] == s {
			b.i = j
			return b
		}
		b.b.Grow(len(b.src))
		b.b.WriteString(b.src[:b.i])
	}
	b.b.WriteString(s)
	return b
}

func (b *lazyBuilder) string() string {
	if b.b.Len() > 0 {
		return b.b.String()
	}
	return b.src[:b.i]
}
