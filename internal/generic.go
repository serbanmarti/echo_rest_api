package internal

import (
	"net/url"
)

// InSlice checks if a string is in a slice
func InSlice(p string, s []string) bool {
	for _, n := range s {
		if p == n {
			return true
		}
	}

	return false
}

// MaxUint gets the maximum of to uint values
func MaxUint(x, y uint) uint {
	if x < y {
		return y
	}

	return x
}

// GetDomainFromURL gets the domain from a URL string
func GetDomainFromURL(u string) (string, error) {
	p, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	return p.Hostname(), nil
}
