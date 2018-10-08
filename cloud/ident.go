// Package cloud provides information about the host cloud platform.
package cloud

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/digitalocean/go-smbios/smbios"
)

// Type is the general cloud type (e.g. AWS, Azure).
type Type int

// Valid cloud types.
const (
	Unknown Type = iota
	AWS
	Azure
)

// String implements fmt.Stringer.
func (t Type) String() string {
	switch t {
	case AWS:
		return "AWS"
	case Azure:
		return "Azure"
	}
	return "Unknown"
}

// Host is the computing resource that is running the code.
type Host int

// Valid host types.
const (
	GenericHost Host = iota
	VM
	Container
	Function
)

// String implements fmt.Stringer.
func (h Host) String() string {
	switch h {
	case VM:
		return "VM"
	case Container:
		return "Container"
	case Function:
		return "Function"
	}
	return "GenericHost"
}

// Info describes the host platform. All fields are populated on a best-effort
// basis.
type Info struct {
	Type   Type
	Host   Host
	HostID string
}

var (
	once sync.Once
	info Info
)

// Ident attempts to identify the cloud platform where the code is running.
func Ident() Info {
	// TODO: Add support for other platforms, handle safe-to-ignore errors
	once.Do(func() {
		r, _, err := smbios.Stream()
		if err != nil {
			return
		}
		defer r.Close()
		structs, err := smbios.NewDecoder(r).Decode()
		if err != nil {
			panic(err)
		}
		// Reference:
		// https://www.dmtf.org/sites/default/files/standards/documents/DSP0134_3.2.0.pdf
		decs := map[uint8]func(*smbios.Structure){
			1: info.decodeSystemInformation,
			3: info.decodeSystemEnclosure,
		}
		for _, s := range structs {
			if dec := decs[s.Header.Type]; dec != nil {
				dec(s)
				if delete(decs, s.Header.Type); len(decs) == 0 {
					break
				}
			}
		}
		return
	})
	return info
}

func (ci *Info) decodeSystemInformation(s *smbios.Structure) {
	id := s.Formatted[4:20]
	if id[0] == 0 || id[0] == 255 {
		for _, b := range id[1:] {
			if b != id[0] {
				goto valid
			}
		}
		return
	valid:
	}
	var (
		timeLow  = binary.LittleEndian.Uint32(id[0:4])
		timeMid  = binary.LittleEndian.Uint16(id[4:6])
		timeHi   = binary.LittleEndian.Uint16(id[6:8])
		clockSeq = id[8:10]
		node     = id[10:16]
	)
	ci.HostID = fmt.Sprintf("%.8x-%.4x-%.4x-%x-%x",
		timeLow, timeMid, timeHi, clockSeq, node)
}

func (ci *Info) decodeSystemEnclosure(s *smbios.Structure) {
	const azureVMAssetTag = "7783-7084-3265-9085-8269-3286-77"
	if i := s.Formatted[4]; i > 0 && s.Strings[i-1] == azureVMAssetTag {
		ci.Type = Azure
		ci.Host = VM
	}
}
