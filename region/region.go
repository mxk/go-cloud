package region

import (
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
)

var (
	once        sync.Once
	regionPart  map[string]string
	partRegions map[string][]string
	svcRegions  map[string][]string
)

// Partition returns the partition of the specified region.
func Partition(region string) string {
	once.Do(load)
	return regionPart[region]
}

// Related returns all regions in the same partition as region.
func Related(region string) []string {
	once.Do(load)
	if rel := partRegions[regionPart[region]]; len(rel) > 0 {
		cpy := make([]string, len(rel))
		copy(cpy, rel)
		return cpy
	}
	return nil
}

// Supports returns true if service is supported in region. Matching is strict,
// so services like IAM are considered to be only in the aws-global region.
func Supports(region, service string) bool {
	once.Do(load)
	return contains(svcRegions[service], region) ||
		(service == endpoints.Ec2metadataServiceID && region == "aws-global")
}

// load creates partition/region/service maps.
func load() {
	parts := endpoints.DefaultPartitions()
	regionPart = make(map[string]string)
	partRegions = make(map[string][]string, len(parts))
	svcRegions = make(map[string][]string)
	for _, p := range parts {
		regionSet := make(map[string]struct{})
		// Using Endpoints() instead of Regions() to handle global services
		for _, s := range p.Services() {
			if s.ID() == endpoints.Ec2metadataServiceID {
				continue
			}
			srs := svcRegions[s.ID()]
			eps := s.Endpoints()
			tmp := make([]string, len(srs), len(srs)+len(eps))
			copy(tmp, srs)
			srs = tmp
			for r := range eps {
				if r != "local" {
					srs = append(srs, r)
					regionSet[r] = struct{}{}
				}
			}
			svcRegions[s.ID()] = srs
		}
		pid := p.ID()
		prs := make([]string, 0, len(regionSet))
		for r := range regionSet {
			if _, dup := regionPart[r]; dup {
				panic("region: duplicate name: " + r)
			}
			regionPart[r] = pid
			prs = append(prs, r)
		}
		sort.Strings(prs)
		partRegions[pid] = prs
	}
	for _, sr := range svcRegions {
		sort.Strings(sr)
	}
}

// contains returns true if v contains s.
func contains(v []string, s string) bool {
	i := sort.SearchStrings(v, s)
	return i < len(v) && v[i] == s
}
