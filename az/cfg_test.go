package az

import (
	"path/filepath"
	"testing"

	"github.com/Azure/go-autorest/version"
	"github.com/LuminalHQ/cloudcover/x/gomod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAuthFile(t *testing.T) {
	goAutorest := gomod.Root(version.UserAgent).Path()
	dir := filepath.Join(goAutorest, "autorest", "azure", "auth", "testdata")
	files := []string{"credsutf8.json", "credsutf16le.json", "credsutf16be.json"}
	want := authFile{
		ClientID:                       "client-id-123",
		ClientSecret:                   "client-secret-456",
		SubscriptionID:                 "sub-id-789",
		TenantID:                       "tenant-id-123",
		ActiveDirectoryEndpointURL:     "https://login.microsoftonline.com",
		ResourceManagerEndpointURL:     "https://management.azure.com/",
		ActiveDirectoryGraphResourceID: "https://graph.windows.net/",
		SQLManagementEndpointURL:       "https://management.core.windows.net:8443/",
		GalleryEndpointURL:             "https://gallery.azure.com/",
		ManagementEndpointURL:          "https://management.core.windows.net/",
	}
	for _, file := range files {
		af, err := loadAuthFile(filepath.Join(dir, file))
		require.NoError(t, err, "%s", file)
		assert.Equal(t, want, af, "%s", file)
	}
}
