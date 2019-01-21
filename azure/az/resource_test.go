package az

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestValidRID(t *testing.T) {
	tests := []*struct {
		id   RID
		norm RID
		rg   string
		prov string
		typ  string
		name string
		get  map[string]string
	}{{
		id: "",
	}, {
		id:   "/subscriptions/" + NilGUID,
		typ:  "subscriptions",
		name: NilGUID,
	}, {
		id:   "/subscriptions/" + NilGUID + "/resourceGroups/test-rg",
		rg:   "test-rg",
		typ:  "resourceGroups",
		name: "test-rg",
	}, {
		id:   "/subscriptions/" + NilGUID + "/providers/Microsoft.Compute",
		prov: "Microsoft.Compute",
		typ:  "providers",
		name: "Microsoft.Compute",
	}, {
		id:   "/subscriptions/" + NilGUID + "/resourcegroups/TEST-RG/providers/Microsoft.Network/virtualNetworks/TEST-VNET",
		norm: "/subscriptions/" + NilGUID + "/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/TEST-VNET",
		rg:   "TEST-RG",
		prov: "Microsoft.Network",
		typ:  "virtualNetworks",
		name: "TEST-VNET",
	}, {
		id:   "/subscriptions/" + NilGUID + "/resourceGroups/test-rg/providers/Microsoft.Network/networkSecurityGroups/test-nsg/securityRules/Port_80",
		rg:   "test-rg",
		prov: "Microsoft.Network",
		typ:  "networkSecurityGroups/securityRules",
		name: "Port_80",
		get: map[string]string{
			"":                      "",
			"Microsoft.Network":     "",
			"networksecuritygroups": "test-nsg",
			"securityRules":         "Port_80",
		},
	}}
	for _, tc := range tests {
		r := tc.id
		if r != "" {
			want := tc.norm
			if want == "" {
				want = r
			}
			norm := r.Norm()
			assert.Equal(t, want, norm, "%+v", tc)
			assert.Equal(t, r == want, noCopy(string(r), string(norm)), "%+v", tc)
			assert.Equal(t, NilGUID, r.Subscription(), "%+v", tc)
		} else {
			assert.Panics(t, func() { r.Norm() })
			assert.Equal(t, "", r.Subscription(), "%+v", tc)
		}
		assert.Equal(t, tc.rg, r.ResourceGroup(), "%+v", tc)
		assert.Equal(t, tc.prov, r.Provider(), "%+v", tc)
		assert.Equal(t, tc.typ, r.Type(), "%+v", tc)
		assert.Equal(t, tc.name, r.Name(), "%+v", tc)
		for k, v := range tc.get {
			assert.Equal(t, v, r.Get(k), "%+v", tc)
		}
	}
}

func TestInvalidRID(t *testing.T) {
	tests := []RID{
		"/",
		"_",
		"/resourceGroup/test-rg",
		"/subscriptions/" + NilGUID + "/",
		"/subscriptions/" + NilGUID + "/resourceGroups",
		"/subscriptions/" + NilGUID + "/resourceGroups/test-rg/virtualNetworks/test-vnet",
		"/subscriptions/" + NilGUID + "/resourceGroups/test-rg/providers/Microsoft.Network",
		"/subscriptions/" + NilGUID + "/providers/Microsoft.Compute/virtualNetworks/test-vnet",
	}
	for _, r := range tests {
		assert.Panics(t, func() { r.Norm() }, "%s", r)
	}
}

func noCopy(a, b string) bool {
	ah := (*reflect.StringHeader)(unsafe.Pointer(&a))
	bh := (*reflect.StringHeader)(unsafe.Pointer(&b))
	return *ah == *bh
}
