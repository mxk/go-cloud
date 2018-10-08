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

// Type returns the resource type.
func (r RID) Type() string {
	v, tail := get(string(r), keyProv)
	if v != "" {
		v, _ = next(tail)
	}
	return v
}

// Name returns the last component of the resource ID.
func (r RID) Name() string {
	return string(r[strings.LastIndexByte(string(r), '/')+1:])
}

// Norm normalizes resource ID representation.
func (r RID) Norm() RID {
	k, sub, tail := nextKeyVal(string(r))
	if !keyEq(k, keySub) {
		r.invalid()
	}
	var rg, prov string
	if tail != "" {
		k, tail = next(tail)
		// Either resource group or provider may be missing, but not both
		if keyEq(k, keyRG) {
			rg, tail = next(tail)
			if tail != "" {
				k, tail = next(tail)
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

// next returns the next component of a resource ID.
func next(r string) (next, tail string) {
	if r == "" || r[0] != '/' {
		tail = r
	} else if i := strings.IndexByte(r[1:], '/') + 1; i == 0 {
		next = r[1:]
	} else {
		next, tail = r[1:i], r[i:]
	}
	return
}

// nextKeyVal returns the next key value pair of a resource ID.
func nextKeyVal(r string) (key, val, tail string) {
	key, tail = next(r)
	val, tail = next(tail)
	return
}

// get returns the value of a resource ID component.
func get(r, key string) (val, tail string) {
	var k string
	for r != "" {
		n := len(r)
		k, r = next(r)
		if keyEq(k, key) {
			return next(r)
		}
		if len(r) == n {
			panic("az: malformed resource id: " + r)
		}
	}
	return
}

// keyEq does case-insensitive comparison of ASCII strings.
func keyEq(s, t string) bool {
	return len(s) == len(t) && (s == t || strings.EqualFold(s, t))
}
