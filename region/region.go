package scan

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
)

// Maps of valid regions for each partition and service. These include global
// regions, such as "aws-global" (see endpoints/defaults.go).
var (
	partRegions = make(map[string]map[string]struct{})
	svcRegions  = make(map[string]map[string]struct{})
	allRegions  = make(map[string]struct{})
)

func init() {
	for _, p := range endpoints.DefaultPartitions() {
		pr := make(map[string]struct{})
		partRegions[p.ID()] = pr
		for _, s := range p.Services() {
			sr := svcRegions[s.ID()]
			if sr == nil {
				sr = make(map[string]struct{})
				svcRegions[s.ID()] = sr
			}
			for region := range s.Endpoints() {
				// TODO: Local DynamoDB scanning requires timeouts, etc.
				if region != "local" {
					sr[region] = struct{}{}
					pr[region] = struct{}{}
					allRegions[region] = struct{}{}
				}
			}
		}
	}
}

// RelatedRegions returns all regions in the same partition as region.
func RelatedRegions(region string) []string {
	if region == "" {
		region = endpoints.UsEast1RegionID
	}
	for _, r := range partRegions {
		if _, ok := r[region]; ok {
			all := make([]string, 0, len(r))
			for name := range r {
				all = append(all, name)
			}
			sort.Strings(all)
			return all
		}
	}
	panic("scan: invalid region: " + region)
}
