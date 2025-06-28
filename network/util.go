package network

import (
	"wispy-core/common"
)

// CreateDomainListFromMap creates a network.DomainList from a map of domains
func CreateDomainListFromMap(domainsMap map[string]string, defaultDomains []string) DomainList {
	domains := NewDomainList()

	// Add all domains from the map
	for domain := range domainsMap {
		if err := domains.AddDomain(domain); err != nil {
			common.Warning("Failed to add domain %q: %v", domain, err)
		}
	}

	// Add default domains
	for _, domain := range defaultDomains {
		if err := domains.AddDomain(domain); err != nil {
			common.Warning("Failed to add default domain %q: %v", domain, err)
		}
	}

	return domains
}
