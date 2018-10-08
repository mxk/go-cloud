package cloud

import (
	"testing"

	"github.com/digitalocean/go-smbios/smbios"
	"github.com/stretchr/testify/assert"
)

func TestDecodeSystemInformation(t *testing.T) {
	var ci Info
	s := &smbios.Structure{
		Header:    smbios.Header{Type: 0x1, Length: 0x19, Handle: 0x1},
		Formatted: []byte{0x1, 0x2, 0x3, 0x4, 0x76, 0x4a, 0xb8, 0x26, 0xa0, 0x1f, 0xd, 0x47, 0xaf, 0x4b, 0x6b, 0xdc, 0x2a, 0xb4, 0xf8, 0xcd, 0x6},
		Strings:   []string{"Microsoft Corporation", "Virtual Machine", "7.0", "0000-0002-7900-6184-0353-2162-69"},
	}
	ci.decodeSystemInformation(s)
	assert.Equal(t, Info{HostID: "26b84a76-1fa0-470d-af4b-6bdc2ab4f8cd"}, ci)

	copy(s.Formatted[4:20], []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	ci = Info{}
	ci.decodeSystemInformation(s)
	assert.Equal(t, Info{HostID: "00000100-0000-0000-0000-000000000000"}, ci)

	s.Formatted[5] = 0
	ci = Info{}
	ci.decodeSystemInformation(s)
	assert.Equal(t, Info{}, ci)
}

func TestDecodeSystemEnclosure(t *testing.T) {
	var ci Info
	ci.decodeSystemEnclosure(&smbios.Structure{
		Header:    smbios.Header{Type: 0x3, Length: 0x11, Handle: 0x3},
		Formatted: []byte{0x1, 0x3, 0x2, 0x3, 0x4, 0x3, 0x3, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0},
		Strings:   []string{"Microsoft Corporation", "7.0", "3127-0109-5694-5247-8764-1345-10", "7783-7084-3265-9085-8269-3286-77"},
	})
	assert.Equal(t, Info{Type: Azure, Host: VM}, ci)
}
