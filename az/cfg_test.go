package az

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/Azure/go-autorest/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAuthFile(t *testing.T) {
	dir := filepath.Join(goAutorestDir(), "testdata")
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

func goAutorestDir() string {
	fn := runtime.FuncForPC(reflect.ValueOf(version.UserAgent).Pointer())
	dir, _ := fn.FileLine(fn.Entry())
	dir = filepath.Dir(filepath.Dir(dir))
	if !strings.HasPrefix(filepath.Base(dir), "go-autorest") {
		panic("go-autorest directory not found")
	}
	return dir
}
