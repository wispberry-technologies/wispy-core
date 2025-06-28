package security

import (
	"fmt"
	"wispy-core/common"
)

type domainsList struct {
	domains *common.HashSet[string]
}

type DomainList interface {
	AddDomain(domain string) error
	HasDomain(domain string) bool
	ListDomains() []string
	Length() int
}

func NewDomainList() DomainList {
	return &domainsList{
		domains: common.NewHashSet[string](),
	}
}
func (d *domainsList) AddDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	if d.domains.Contains(domain) {
		return fmt.Errorf("domain already exists: %s", domain)
	}
	d.domains.Add(domain)
	return nil
}
func (d *domainsList) HasDomain(domain string) bool {
	if domain == "" {
		return false
	}
	return d.domains.Contains(domain)
}
func (d *domainsList) ListDomains() []string {
	output := []string{}
	for domain := range d.domains.All() {
		if domain == "" {
			continue
		}
		output = append(output, domain)
	}
	return output
}
func (d *domainsList) Length() int {
	return d.domains.Len()
}
