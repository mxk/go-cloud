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
		name string
	}{{
		id: "",
	}, {
		id:   "/subscriptions/" + guid,
		typ:  "subscriptions",
		name: guid,
	}, {
		id:   "/subscriptions/" + guid + "/resourceGroups/test-rg",
		rg:   "test-rg",
		typ:  "resourceGroups",
		name: "test-rg",
	}, {
		id:   "/subscriptions/" + guid + "/providers/Microsoft.Compute",
		prov: "Microsoft.Compute",
		typ:  "providers",
		name: "Microsoft.Compute",
	}, {
		id:   "/subscriptions/" + guid + "/resourcegroups/TEST-RG/providers/Microsoft.Network/virtualNetworks/TEST-VNET",
		norm: "/subscriptions/" + guid + "/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/TEST-VNET",
		rg:   "TEST-RG",
		prov: "Microsoft.Network",
		typ:  "Microsoft.Network/virtualNetworks",
		name: "TEST-VNET",
	}, {
		id:   "/subscriptions/" + guid + "/resourceGroups/test-rg/providers/Microsoft.Network/networkSecurityGroups/test-nsg/securityRules/Port_80",
		rg:   "test-rg",
		prov: "Microsoft.Network",
		typ:  "Microsoft.Network/networkSecurityGroups/securityRules",
		name: "Port_80",
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
		assert.Equal(t, tc.name, r.Name(), "%+v", tc)
	}
}
