package az

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	imdsMetadata = "http://169.254.169.254/metadata"
	imdsVersion  = "2018-02-01"
)

// TODO: Update types and add missing fields

// InstanceMetadata contains VM metadata obtained from the Instance Metadata
// Service (IMDS).
type InstanceMetadata struct {
	Compute ComputeMetadata
}

// ComputeMetadata contains basic VM information.
type ComputeMetadata struct {
	Location             string
	Name                 string
	Offer                string
	OSType               string
	PlacementGroupID     string
	PlatformFaultDomain  string
	PlatformUpdateDomain string
	Publisher            string
	ResourceGroupName    string
	SKU                  string
	SubscriptionID       string
	Tags                 string
	Version              string
	VMID                 string
	VMScaleSetName       string
	VMSize               string
	Zone                 string
}

// GetInstanceMetadata returns VM metadata from IMDS.
func GetInstanceMetadata() (im InstanceMetadata, err error) {
	req, err := http.NewRequest(http.MethodGet,
		imdsMetadata+"/instance?api-version="+imdsVersion, nil)
	if err != nil {
		return
	}
	req.Header.Set("Metadata", "true")
	rsp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		err = fmt.Errorf("az: failed to retrieve instance metadata (%s)",
			rsp.Status)
		return
	}
	var buf bytes.Buffer
	if rsp.ContentLength > 0 {
		buf.Grow(int(rsp.ContentLength))
	}
	if _, err = buf.ReadFrom(rsp.Body); err == nil {
		err = json.Unmarshal(buf.Bytes(), &im)
	}
	return
}
