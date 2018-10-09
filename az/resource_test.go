package az

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRID(t *testing.T) {
	const guid = "00000000-0000-0000-0000-000000000000"
	tests := []*struct {
		id   RID
		norm RID
		rg   string
		prov string
		typ  string
		kind string
		name string
	}{{
		id: "",
	}, {
		id:   "/subscriptions/" + guid,
		kind: "subscriptions",
		name: guid,
	}, {
		id:   "/subscriptions/" + guid + "/resourceGroups/test-rg",
		rg:   "test-rg",
		kind: "resourceGroups",
		name: "test-rg",
	}, {
		id:   "/subscriptions/" + guid + "/providers/Microsoft.Compute",
		prov: "Microsoft.Compute",
		kind: "providers",
		name: "Microsoft.Compute",
	}, {
		id:   "/subscriptions/" + guid + "/resourcegroups/TEST-RG/providers/Microsoft.Network/virtualNetworks/TEST-VNET",
		norm: "/subscriptions/" + guid + "/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/TEST-VNET",
		rg:   "TEST-RG",
		prov: "Microsoft.Network",
		typ:  "virtualNetworks",
		kind: "Microsoft.Network/virtualNetworks",
		name: "TEST-VNET",
	}}
	for _, tc := range tests {
		r := tc.id
		if r != "" {
			if tc.norm == "" {
				tc.norm = r
			}
			assert.Equal(t, tc.norm, r.Norm(), "%+v", tc)
			assert.Equal(t, guid, r.Subscription(), "%+v", tc)
		} else {
			assert.Panics(t, func() { r.Norm() })
			assert.Equal(t, "", r.Subscription(), "%+v", tc)
		}
		assert.Equal(t, tc.rg, r.ResourceGroup(), "%+v", tc)
		assert.Equal(t, tc.prov, r.Provider(), "%+v", tc)
		assert.Equal(t, tc.typ, r.Type(), "%+v", tc)
		assert.Equal(t, tc.kind, r.Kind(), "%+v", tc)
		assert.Equal(t, tc.name, r.Name(), "%+v", tc)
	}
}
