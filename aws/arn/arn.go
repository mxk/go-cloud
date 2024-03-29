// Package arn provides functions and types for working with Amazon Resource
// Names.
package arn

import (
	"path"
	"strconv"
	"strings"
)

const (
	prefix = ARN("arn:")
	fields = 5
	buf    = 256
)

// ARN is an Amazon Resource Name.
type ARN string

// Base is an ARN without any fields set.
const Base = ARN(prefix + "::::")

// New constructs an ARN from the specified fields. Any field may be left blank.
func New(partition, service, region, account string, resource ...string) ARN {
	var buf [buf]byte
	b := append(buf[:0], prefix...)
	b = append(append(b, partition...), ':')
	b = append(append(b, service...), ':')
	b = append(append(b, region...), ':')
	b = append(append(b, account...), ':')
	for _, r := range resource {
		b = append(b, r...)
	}
	return ARN(b)
}

// Value returns the value of the ARN string pointer passed in or "" if the
// pointer is nil.
func Value(s *string) ARN {
	if s == nil {
		return ""
	}
	return ARN(*s)
}

// String returns a pointer to the ARN value passed in.
func String(r ARN) *string {
	s := string(r)
	return &s
}

// Partition returns ARN partition.
func (r ARN) Partition() string { return r.Field(0) }

// Service returns ARN service.
func (r ARN) Service() string { return r.Field(1) }

// Region returns ARN region.
func (r ARN) Region() string { return r.Field(2) }

// Account returns ARN account.
func (r ARN) Account() string { return r.Field(3) }

// Resource returns ARN resource.
func (r ARN) Resource() string { return r.Field(4) }

// WithPartition returns a new ARN with partition replaced by v.
func (r ARN) WithPartition(v string) ARN { return r.WithField(0, v) }

// WithService returns a new ARN with service replaced by v.
func (r ARN) WithService(v string) ARN { return r.WithField(1, v) }

// WithRegion returns a new ARN with region replaced by v.
func (r ARN) WithRegion(v string) ARN { return r.WithField(2, v) }

// WithAccount returns a new ARN with account replaced by v.
func (r ARN) WithAccount(v string) ARN { return r.WithField(3, v) }

// WithResource returns a new ARN with resource replaced by v.
func (r ARN) WithResource(v string) ARN { return r.WithField(4, v) }

// Valid returns true if r has a valid prefix and the required number of fields.
func (r ARN) Valid() bool {
	return strings.HasPrefix(string(r), string(prefix)) &&
		strings.Count(string(r), ":") >= fields
}

// Type returns the resource prefix up to the first '/' or ':' character. It
// returns an empty string if neither character is found.
func (r ARN) Type() string {
	i, j := r.typ()
	return string(r[i:j])
}

// Path returns the resource substring between and including the first and last
// '/' characters. It ignores any part of the resource before the last ':' and
// returns an empty string if the resource does not contain any '/' characters.
func (r ARN) Path() string {
	i, j := r.path()
	return string(r[i:j])
}

// Name returns the resource suffix after the last '/' or ':' character. It
// returns the whole resource field if neither character is found.
func (r ARN) Name() string {
	return string(r[r.name():])
}

// PathName returns the combined resource path and name. It panics if r does not
// have a path.
func (r ARN) PathName() string {
	i, j := r.path()
	if i == j {
		panic("arn: no path: " + string(r))
	}
	return string(r[i:])
}

// WithPath returns a new ARN with path replaced by v. It panics if r does not
// have a path.
func (r ARN) WithPath(v string) ARN {
	i, j := r.path()
	if i == j {
		panic("arn: no path: " + string(r))
	}
	return concat(r[:i], cleanPath(v), "/", r[j:])
}

// WithName returns a new ARN with name replaced by v.
func (r ARN) WithName(v string) ARN {
	return concat(r[:r.name()], ARN(v))
}

// WithPathName returns a new ARN with path and name replaced by v. It panics if
// r does not have a path.
func (r ARN) WithPathName(v string) ARN {
	i, j := r.path()
	if i == j {
		panic("arn: no path: " + string(r))
	}
	j = strings.LastIndexByte(v, '/')
	return concat(r[:i], cleanPath(v[:j+1]), "/", ARN(v[j+1:]))
}

// Field returns the ith ARN field.
func (r ARN) Field(i int) string {
	j, k := r.field(i)
	return string(r[j:k])
}

// WithField returns a new ARN with the ith field set to v.
func (r ARN) WithField(i int, v string) ARN {
	j, k := r.field(i)
	if string(r[j:k]) == v {
		return r
	}
	return concat(r[:j], ARN(v), r[k:])
}

// With returns a new ARN, with non-empty fields in o replacing those in r.
func (r ARN) With(o ARN) ARN {
	var f [fields]string
	for i := 0; i < fields; i++ {
		if s := o.Field(i); s != "" {
			f[i] = s
		} else {
			f[i] = r.Field(i)
		}
	}
	return New(f[0], f[1], f[2], f[3], f[4])
}

// Ctx extracts the partition, region, and account fields from r.
func (r ARN) Ctx() Ctx {
	return Ctx{r.Partition(), r.Region(), r.Account()}
}

// field returns slice indices of the ith field.
func (r ARN) field(i int) (int, int) {
	n, j, k := i, len(prefix), len(prefix)
	if len(r) < j || r[:j] != prefix {
		panic("arn: invalid arn: " + string(r))
	}
	for ; k < len(r); k++ {
		if r[k] == ':' {
			if n <= 1 {
				if n <= 0 {
					return j, k
				} else if i >= fields-1 {
					return k + 1, len(r)
				}
			}
			j = k + 1
			n--
		}
	}
	panic("arn: invalid arn or field index " + strconv.Itoa(i) + ": " +
		string(r))
}

// typ returns the slice indices of the resource type.
func (r ARN) typ() (int, int) {
	i, _ := r.field(fields - 1)
	for j := i; j < len(r); j++ {
		switch r[j] {
		case '/', ':':
			return i, j
		}
	}
	return i, i
}

// path returns slice indices of the resource path.
func (r ARN) path() (int, int) {
	for i, j, k := len(r)-1, 0, 0; i >= 0; i-- {
		switch r[i] {
		case '/':
			if j = i; k == 0 {
				k = i + 1
			}
		case ':':
			return j, k
		}
	}
	panic("arn: invalid arn: " + string(r))
}

// name returns the starting index of the resource name.
func (r ARN) name() int {
	for i := len(r) - 1; i >= 0; i-- {
		switch r[i] {
		case '/', ':':
			return i + 1
		}
	}
	panic("arn: invalid arn: " + string(r))
}

// cleanPath normalizes path p, returning either an empty string or an absolute
// path without a trailing '/'.
func cleanPath(p string) ARN {
	if p != "" {
		if p[0] != '/' {
			p = "/" + p
		}
		if p = path.Clean(p); p == "/" {
			p = ""
		}
	}
	return ARN(p)
}

// concat concatenates all parts of an ARN.
func concat(parts ...ARN) ARN {
	var buf [buf]byte
	b := buf[:0]
	for _, s := range parts {
		b = append(b, s...)
	}
	return ARN(b)
}
