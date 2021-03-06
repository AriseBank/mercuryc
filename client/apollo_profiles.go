package apollo

import (
	"fmt"
	"strings"

	"github.com/AriseBank/apollo-controller/shared/api"
)

// Profile handling functions

// GetProfileNames returns a list of available profile names
func (r *ProtocolAPOLLO) GetProfileNames() ([]string, error) {
	urls := []string{}

	// Fetch the raw value
	_, err := r.queryStruct("GET", "/profiles", nil, "", &urls)
	if err != nil {
		return nil, err
	}

	// Parse it
	names := []string{}
	for _, url := range urls {
		fields := strings.Split(url, "/profiles/")
		names = append(names, fields[len(fields)-1])
	}

	return names, nil
}

// GetProfiles returns a list of available Profile structs
func (r *ProtocolAPOLLO) GetProfiles() ([]api.Profile, error) {
	profiles := []api.Profile{}

	// Fetch the raw value
	_, err := r.queryStruct("GET", "/profiles?recursion=1", nil, "", &profiles)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

// GetProfile returns a Profile entry for the provided name
func (r *ProtocolAPOLLO) GetProfile(name string) (*api.Profile, string, error) {
	profile := api.Profile{}

	// Fetch the raw value
	etag, err := r.queryStruct("GET", fmt.Sprintf("/profiles/%s", name), nil, "", &profile)
	if err != nil {
		return nil, "", err
	}

	return &profile, etag, nil
}

// CreateProfile defines a new container profile
func (r *ProtocolAPOLLO) CreateProfile(profile api.ProfilesPost) error {
	// Send the request
	_, _, err := r.query("POST", "/profiles", profile, "")
	if err != nil {
		return err
	}

	return nil
}

// UpdateProfile updates the profile to match the provided Profile struct
func (r *ProtocolAPOLLO) UpdateProfile(name string, profile api.ProfilePut, ETag string) error {
	// Send the request
	_, _, err := r.query("PUT", fmt.Sprintf("/profiles/%s", name), profile, ETag)
	if err != nil {
		return err
	}

	return nil
}

// RenameProfile renames an existing profile entry
func (r *ProtocolAPOLLO) RenameProfile(name string, profile api.ProfilePost) error {
	// Send the request
	_, _, err := r.query("POST", fmt.Sprintf("/profiles/%s", name), profile, "")
	if err != nil {
		return err
	}

	return nil
}

// DeleteProfile deletes a profile
func (r *ProtocolAPOLLO) DeleteProfile(name string) error {
	// Send the request
	_, _, err := r.query("DELETE", fmt.Sprintf("/profiles/%s", name), nil, "")
	if err != nil {
		return err
	}

	return nil
}
