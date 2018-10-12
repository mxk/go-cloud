package az

import (
	"strings"
)

const (
	keySub  = "subscriptions"
	keyRG   = "resourceGroups"
	keyProv = "providers"
)

// RID is an Azure resource ID.
type RID string

// Subscription returns the resource subscription ID.
func (r RID) Subscription() string {
	v, _ := get(string(r), keySub)
	return v
}

// ResourceGroup returns the name of the resource group.
func (r RID) ResourceGroup() string {
	v, _ := get(string(r), keyRG)
	return v
}

// Provider returns the name of the resource provider.
func (r RID) Provider() string {
	v, _ := get(string(r), keyProv)
	return v
}

// Type returns the final key component of r (e.g. "resourceGroups") or
// "<provider>/<type>[/<subtype>...]" if r contains a "/providers/" component.
func (r RID) Type() string {
	tail := string(r)
	for {
		k, _, t := nextPair(tail)
		if t == "" {
			return k
		}
		if k == keyProv {
			tail = tail[1+len(keyProv):]
			break
		}
		tail = t
	}
	// Extract "<provider>/<type>" substring without allocation
	i, j := nextIdx(tail)
	_, l := nextIdx(tail[j:])
	j += l
	k := tail[i:j]
	if _, tail = nextStr(tail[j:]); tail != "" {
		// Concatenate subtypes
		var b strings.Builder
		b.WriteString(k)
		for tail != "" {
			k, _, tail = nextPair(tail)
			b.WriteByte('/')
			b.WriteString(k)
		}
		k = b.String()
	}
	return k
}

// Name returns the last component of the resource ID (path basename).
func (r RID) Name() string {
	return string(r[strings.LastIndexByte(string(r), '/')+1:])
}

// Norm normalizes resource ID representation.
func (r RID) Norm() RID {
	k, sub, tail := nextPair(string(r))
	if !keyEq(k, keySub) {
		r.invalid()
	}
	var rg, prov string
	if tail != "" {
		k, tail = nextStr(tail)
		// Either resource group or provider may be missing, but not both
		if keyEq(k, keyRG) {
			rg, tail = nextStr(tail)
			if tail != "" {
				k, tail = nextStr(tail)
			} else {
				k = keyProv
			}
		}
		if !keyEq(k, keyProv) {
			r.invalid()
		}
		prov = tail
	}
	var b strings.Builder
	b.Grow(len(r))
	b.WriteString("/" + keySub + "/")
	b.WriteString(strings.ToLower(sub))
	if rg != "" {
		b.WriteString("/" + keyRG + "/")
		b.WriteString(strings.ToLower(rg))
	}
	if prov != "" {
		// TODO: Normalize provider name?
		b.WriteString("/" + keyProv)
		b.WriteString(prov)
	}
	if norm := RID(b.String()); norm != r {
		return norm
	}
	return r
}

// invalid panics to indicate malformed resource ID.
func (r RID) invalid() {
	panic("az: invalid resource id: " + string(r))
}

// get returns the value of a resource ID component.
func get(r, key string) (val, tail string) {
	var k, v string
	for r != "" {
		if k, v, r = nextPair(r); keyEq(k, key) {
			return v, r
		}
	}
	return
}

// nextPair returns the next resource ID key/value pair.
func nextPair(r string) (k, v, tail string) {
	if r != "" {
		i, j := nextIdx(r)
		if k, r = r[i:j], r[j:]; r == "" {
			panic("az: missing component value in resource id: " + r)
		}
		i, j = nextIdx(r)
		v, tail = r[i:j], r[j:]
	}
	return
}

// nextStr returns the next resource ID component as a string.
func nextStr(r string) (s, tail string) {
	i, j := nextIdx(r)
	return r[i:j], r[j:]
}

// nextIdx returns the indices of the next resource ID component. It panics if a
// non-empty r does not begin with a '/'.
func nextIdx(r string) (i, j int) {
	if r != "" {
		if r[0] != '/' {
			panic("az: malformed resource id: " + r)
		}
		if i, j = 1, strings.IndexByte(r[1:], '/')+1; j == 0 {
			j = len(r)
		}
	}
	return
}

// keyEq does case-insensitive comparison of ASCII strings.
func keyEq(s, t string) bool {
	return len(s) == len(t) && (s == t || strings.EqualFold(s, t))
}
