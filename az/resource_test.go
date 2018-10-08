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
		id:   "/subscriptions/" + guid,
		name: guid,
	}, {
		id:   "/subscriptions/" + guid + "/resourceGroups/test-rg",
		rg:   "test-rg",
		name: "test-rg",
	}, {
		id:   "/subscriptions/" + guid + "/providers/Microsoft.Compute",
		prov: "Microsoft.Compute",
		name: "Microsoft.Compute",
	}, {
		id:   "/subscriptions/" + guid + "/resourcegroups/TEST-RG/providers/Microsoft.Network/virtualNetworks/TEST-VNET",
		norm: "/subscriptions/" + guid + "/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/TEST-VNET",
		rg:   "TEST-RG",
		prov: "Microsoft.Network",
		typ:  "virtualNetworks",
		name: "TEST-VNET",
	}}
	for _, tc := range tests {
		r := tc.id
		norm := r.Norm()
		want := tc.norm
		if want == "" {
			want = r
		}
		assert.Equal(t, want, norm, "%+v", tc)
		assert.Equal(t, guid, r.Subscription(), "%+v", tc)
		assert.Equal(t, tc.rg, r.ResourceGroup(), "%+v", tc)
		assert.Equal(t, tc.prov, r.Provider(), "%+v", tc)
		assert.Equal(t, tc.typ, r.Type(), "%+v", tc)
		assert.Equal(t, tc.name, r.Name(), "%+v", tc)
	}
}
