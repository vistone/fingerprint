package gateway

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/errors"
	"github.com/vistone/fingerprint/modules/profiles"
)

func (pm *ProfileManager) GetProfile(id string) (*profiles.ClientProfile, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	profile, ok := pm.profiles[id]
	if !ok {
		return nil, errors.ProfileNotFound(id)
	}
	clone := *profile
	return &clone, nil
}

// GetDefaultProfile returns the default Profile
func (pm *ProfileManager) GetDefaultProfile() (*profiles.ClientProfile, error) {
	return pm.GetProfile(pm.defaultID)
}

// SetDefaultProfile sets the default Profile ID
func (pm *ProfileManager) SetDefaultProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, ok := pm.profiles[id]; !ok {
		return errors.ProfileNotFound(id)
	}

	pm.defaultID = id
	return nil
}

// ListProfiles lists all Profile IDs
func (pm *ProfileManager) ListProfiles() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	ids := make([]string, 0, len(pm.profiles))
	for id := range pm.profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// AddProfile dynamically adds a Profile
func (pm *ProfileManager) AddProfile(profile *profiles.ClientProfile) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if profile.ID == "" {
		return errors.RequiredField("profile.ID")
	}

	pm.profiles[profile.ID] = profile
	return nil
}

// RemoveProfile removes a Profile
func (pm *ProfileManager) RemoveProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if id == pm.defaultID {
		return errors.ProfileRemoveDefault(id)
	}

	delete(pm.profiles, id)
	return nil
}

// SaveProfile saves a Profile to file
func (pm *ProfileManager) SaveProfile(id string) error {
	pm.mu.RLock()
	profile, ok := pm.profiles[id]
	pm.mu.RUnlock()

	if !ok {
		return errors.ProfileNotFound(id)
	}

	// Ensure directory exists
	if err := os.MkdirAll(pm.configDir, 0755); err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileSaveFailed,
			"failed to create config dir", err).WithDetail("profile_id", id)
	}

	// Serialize Profile
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileSaveFailed,
			"failed to marshal profile", err).WithDetail("profile_id", id)
	}

	// Write to file
	filename := filepath.Join(pm.configDir, fmt.Sprintf("%s.json", id))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileSaveFailed,
			"failed to write profile file", err).WithDetail("profile_id", id)
	}

	return nil
}

// ExportProfilesExample exports example configuration files (for documentation and initialization)
func (pm *ProfileManager) ExportProfilesExample() error {
	// Load default configuration
	if err := pm.LoadDefaultProfiles(); err != nil {
		return err
	}

	// Export all default configurations
	for id := range pm.profiles {
		if err := pm.SaveProfile(id); err != nil {
			return err
		}
	}

	return nil
}

// ReloadProfile reloads the specified Profile
func (pm *ProfileManager) ReloadProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if it's a built-in profile
	if _, ok := pm.profiles[id]; !ok {
		return errors.ProfileNotFound(id)
	}

	// Try to reload from file
	filename := filepath.Join(pm.configDir, fmt.Sprintf("%s.json", id))
	if _, err := os.Stat(filename); err != nil {
		// File doesn't exist, return current in-memory version
		return nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileLoadFailed,
			"failed to read profile file", err).WithDetail("profile_id", id)
	}

	var profile profiles.ClientProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeProfileInvalid,
			"failed to parse profile", err).WithDetail("profile_id", id)
	}

	pm.profiles[id] = &profile
	return nil
}

// ReloadAll reloads all Profiles
func (pm *ProfileManager) ReloadAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Reload built-in profiles
	allProfiles := profiles.GetAll()
	for _, p := range allProfiles {
		profile := p
		pm.profiles[profile.ID] = &profile
	}

	// Load local profile files from config directory
	if _, err := os.Stat(pm.configDir); !os.IsNotExist(err) {
		files, _ := filepath.Glob(filepath.Join(pm.configDir, "*.json"))
		for _, file := range files {
			if data, err := os.ReadFile(file); err == nil {
				var profile profiles.ClientProfile
				if err := json.Unmarshal(data, &profile); err == nil && profile.ID != "" {
					pm.profiles[profile.ID] = &profile
				}
			}
		}
	}

	return nil
}

// GetProfilesByBrowser returns Profiles by browser type
func (pm *ProfileManager) GetProfilesByBrowser(browser core.BrowserType) []*profiles.ClientProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []*profiles.ClientProfile
	for _, p := range pm.profiles {
		if p.BrowserType == browser {
			result = append(result, p)
		}
	}
	return result
}

// GetProfilesByOS returns Profiles by operating system
func (pm *ProfileManager) GetProfilesByOS(os core.OperatingSystem) []*profiles.ClientProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []*profiles.ClientProfile
	for _, p := range pm.profiles {
		if p.OS == os {
			result = append(result, p)
		}
	}
	return result
}

// Count returns the number of Profiles
func (pm *ProfileManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.profiles)
}

// CloneProfile clones a Profile
func (pm *ProfileManager) CloneProfile(sourceID, newID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if newID == "" {
		return errors.RequiredField("newID")
	}

	if _, exists := pm.profiles[newID]; exists {
		return fmt.Errorf("profile ID %q already exists", newID)
	}

	source, ok := pm.profiles[sourceID]
	if !ok {
		return errors.ProfileNotFound(sourceID)
	}

	// Create a copy
	clone := *source
	clone.ID = newID
	// Modify name to distinguish
	clone.Name = fmt.Sprintf("%s (Clone)", source.Name)

	pm.profiles[newID] = &clone
	return nil
}
