// Package config provides configuration types and loading functionality.
package config

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

const (
	// NameModeName keeps the configured resource names unchanged.
	NameModeName = "name"
	// NameModeSuffix appends a generated unique suffix to configured resource names.
	NameModeSuffix = "suffix"
	// defaultRandomSuffixLength controls the random tail length for generated suffixes.
	defaultRandomSuffixLength = 4
)

var (
	// shortSuffixAlphabet contains safe lowercase alphanumeric characters for generated suffixes.
	shortSuffixAlphabet = []byte("abcdefghijklmnopqrstuvwxyz0123456789")
	// randomRead allows tests to exercise fallback branches without mutating crypto/rand internals.
	randomRead = rand.Read
)

// ResolveNames resolves nameMode for all supported resources and rewrites references.
// It returns the resolved configuration and the generated suffix.
// The returned suffix is empty when no suffix mode is applied.
func (c *Config) ResolveNames() (*Config, string, error) {
	suffix := GenerateTimestampSuffix()
	resolved, applied, err := c.resolveNamesWithSuffix(suffix)
	if err != nil {
		return nil, "", err
	}
	if !applied {
		return resolved, "", nil
	}
	return resolved, suffix, nil
}

// ResolveNamesWithSuffix resolves nameMode with a caller-provided suffix.
// This helper is mainly intended for deterministic tests.
func (c *Config) ResolveNamesWithSuffix(suffix string) (*Config, error) {
	resolved, _, err := c.resolveNamesWithSuffix(suffix)
	if err != nil {
		return nil, err
	}
	return resolved, nil
}

// GenerateTimestampSuffix returns a unique suffix in yyyyMMddHHmmssSSS-rand4 format.
func GenerateTimestampSuffix() string {
	now := time.Now()
	millisecond := now.Nanosecond() / int(time.Millisecond)
	return fmt.Sprintf("%s%03d-%s", now.Format("20060102150405"), millisecond, generateShortRandomSuffix(defaultRandomSuffixLength))
}

// resolveNamesWithSuffix performs a full nameMode resolution and returns whether suffix mode was applied.
func (c *Config) resolveNamesWithSuffix(suffix string) (*Config, bool, error) {
	if c == nil {
		return &Config{}, false, nil
	}

	defaultMode, err := normalizeNameMode(c.NameMode, NameModeName)
	if err != nil {
		return nil, false, fmt.Errorf("invalid config.nameMode: %w", err)
	}
	// Prevalidate suffix once to avoid repeated checks in every resource loop.
	if defaultMode == NameModeSuffix && suffix == "" {
		return nil, false, fmt.Errorf("nameMode suffix requires a non-empty suffix")
	}

	resolved := cloneConfig(c)
	appliedSuffix := false

	userIDMap := make(map[string]string, len(resolved.Users))
	repositoryNameMap := make(map[string]string, len(resolved.Repositories))
	roleIDMap := make(map[string]string, len(resolved.Roles))
	privilegeNameMap := make(map[string]string, len(resolved.Privileges))

	for i, user := range resolved.Users {
		mode, err := normalizeNameMode(user.NameMode, defaultMode)
		if err != nil {
			return nil, false, fmt.Errorf("invalid users[%d].nameMode: %w", i, err)
		}
		if mode == NameModeSuffix && suffix == "" {
			return nil, false, fmt.Errorf("nameMode suffix requires a non-empty suffix")
		}

		actualID := user.ID
		if mode == NameModeSuffix {
			appliedSuffix = true
			actualID = withSuffix(user.ID, suffix)
			resolved.Users[i].ID = actualID
			resolved.Users[i].EmailAddress = emailWithSuffix(user.EmailAddress, suffix)
		}
		userIDMap[user.ID] = actualID
	}

	for i, repository := range resolved.Repositories {
		mode, err := normalizeNameMode(repository.NameMode, defaultMode)
		if err != nil {
			return nil, false, fmt.Errorf("invalid repositories[%d].nameMode: %w", i, err)
		}
		if mode == NameModeSuffix && suffix == "" {
			return nil, false, fmt.Errorf("nameMode suffix requires a non-empty suffix")
		}

		actualName := repository.Name
		if mode == NameModeSuffix {
			appliedSuffix = true
			actualName = withSuffix(repository.Name, suffix)
			resolved.Repositories[i].Name = actualName
		}
		repositoryNameMap[repository.Name] = actualName
	}

	for i, privilege := range resolved.Privileges {
		mode, err := normalizeNameMode(privilege.NameMode, defaultMode)
		if err != nil {
			return nil, false, fmt.Errorf("invalid privileges[%d].nameMode: %w", i, err)
		}
		if mode == NameModeSuffix && suffix == "" {
			return nil, false, fmt.Errorf("nameMode suffix requires a non-empty suffix")
		}

		actualName := privilege.Name
		if mode == NameModeSuffix {
			appliedSuffix = true
			actualName = withSuffix(privilege.Name, suffix)
			resolved.Privileges[i].Name = actualName
		}
		privilegeNameMap[privilege.Name] = actualName
	}

	for i, role := range resolved.Roles {
		mode, err := normalizeNameMode(role.NameMode, defaultMode)
		if err != nil {
			return nil, false, fmt.Errorf("invalid roles[%d].nameMode: %w", i, err)
		}
		if mode == NameModeSuffix && suffix == "" {
			return nil, false, fmt.Errorf("nameMode suffix requires a non-empty suffix")
		}

		actualID := role.ID
		if mode == NameModeSuffix {
			appliedSuffix = true
			actualID = withSuffix(role.ID, suffix)
			resolved.Roles[i].ID = actualID
			resolved.Roles[i].Name = withSuffix(role.Name, suffix)
		}
		roleIDMap[role.ID] = actualID
	}

	for i := range resolved.Users {
		resolved.Users[i].Roles = remapStrings(resolved.Users[i].Roles, roleIDMap)
	}

	for i := range resolved.Roles {
		resolved.Roles[i].Privileges = remapStrings(resolved.Roles[i].Privileges, privilegeNameMap)
		resolved.Roles[i].Roles = remapStrings(resolved.Roles[i].Roles, roleIDMap)
	}

	for i := range resolved.Privileges {
		if mappedRepo, ok := repositoryNameMap[resolved.Privileges[i].Repository]; ok {
			resolved.Privileges[i].Repository = mappedRepo
		}
	}

	for i := range resolved.UserRepositoryPermissions {
		if mappedUser, ok := userIDMap[resolved.UserRepositoryPermissions[i].UserID]; ok {
			resolved.UserRepositoryPermissions[i].UserID = mappedUser
		}
		if mappedRepo, ok := repositoryNameMap[resolved.UserRepositoryPermissions[i].Repository]; ok {
			resolved.UserRepositoryPermissions[i].Repository = mappedRepo
		}
	}

	clearNameModes(resolved)

	return resolved, appliedSuffix, nil
}

// normalizeNameMode validates and normalizes a nameMode value.
func normalizeNameMode(mode string, fallback string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		normalized = fallback
	}
	switch normalized {
	case NameModeName, NameModeSuffix:
		return normalized, nil
	default:
		return "", fmt.Errorf("unsupported nameMode %q (allowed: %q, %q)", mode, NameModeName, NameModeSuffix)
	}
}

// withSuffix appends suffix in "<value>-<suffix>" format.
func withSuffix(value string, suffix string) string {
	if value == "" || suffix == "" {
		return value
	}
	return fmt.Sprintf("%s-%s", value, suffix)
}

// emailWithSuffix appends suffix to email local-part to avoid collisions.
func emailWithSuffix(email string, suffix string) string {
	if email == "" || suffix == "" {
		return email
	}
	at := strings.Index(email, "@")
	if at < 0 {
		return withSuffix(email, suffix)
	}
	return fmt.Sprintf("%s+%s%s", email[:at], suffix, email[at:])
}

// generateShortRandomSuffix returns a short lowercase alphanumeric suffix.
func generateShortRandomSuffix(length int) string {
	if length <= 0 {
		length = defaultRandomSuffixLength
	}

	rawRandomBytes := make([]byte, length)
	if _, err := randomRead(rawRandomBytes); err != nil {
		fallbackFromTime := fmt.Sprintf("%d", time.Now().UnixNano())
		if len(fallbackFromTime) > length {
			return fallbackFromTime[len(fallbackFromTime)-length:]
		}
		return fallbackFromTime
	}

	randomSuffix := make([]byte, length)
	for i := range rawRandomBytes {
		randomSuffix[i] = shortSuffixAlphabet[int(rawRandomBytes[i])%len(shortSuffixAlphabet)]
	}

	return string(randomSuffix)
}

// remapStrings remaps values using map and keeps unknown values unchanged.
func remapStrings(values []string, mapping map[string]string) []string {
	if len(values) == 0 {
		return values
	}
	remapped := make([]string, 0, len(values))
	for _, value := range values {
		if mapped, ok := mapping[value]; ok {
			remapped = append(remapped, mapped)
			continue
		}
		remapped = append(remapped, value)
	}
	return remapped
}

// clearNameModes removes nameMode fields from a resolved config so it can be safely reused.
func clearNameModes(cfg *Config) {
	if cfg == nil {
		return
	}

	cfg.NameMode = ""
	for i := range cfg.Users {
		cfg.Users[i].NameMode = ""
	}
	for i := range cfg.Repositories {
		cfg.Repositories[i].NameMode = ""
	}
	for i := range cfg.Privileges {
		cfg.Privileges[i].NameMode = ""
	}
	for i := range cfg.Roles {
		cfg.Roles[i].NameMode = ""
	}
}

// cloneConfig creates a deep copy for all mutable slices.
func cloneConfig(in *Config) *Config {
	if in == nil {
		return &Config{}
	}

	out := &Config{
		NameMode:                  in.NameMode,
		Users:                     make([]User, len(in.Users)),
		Repositories:              make([]Repository, len(in.Repositories)),
		Privileges:                make([]Privilege, len(in.Privileges)),
		Roles:                     make([]Role, len(in.Roles)),
		UserRepositoryPermissions: make([]UserRepositoryPermission, len(in.UserRepositoryPermissions)),
	}

	for i, user := range in.Users {
		out.Users[i] = user
		out.Users[i].Roles = copyStrings(user.Roles)
	}
	for i, repository := range in.Repositories {
		out.Repositories[i] = repository
		if repository.Proxy != nil {
			proxyCopy := *repository.Proxy
			if repository.Proxy.Authentication != nil {
				authCopy := *repository.Proxy.Authentication
				proxyCopy.Authentication = &authCopy
			}
			out.Repositories[i].Proxy = &proxyCopy
		}
		if repository.Maven != nil {
			mavenCopy := *repository.Maven
			out.Repositories[i].Maven = &mavenCopy
		}
		if repository.Docker != nil {
			dockerCopy := *repository.Docker
			out.Repositories[i].Docker = &dockerCopy
		}
		if repository.Apt != nil {
			aptCopy := *repository.Apt
			out.Repositories[i].Apt = &aptCopy
		}
		if repository.Cleanup != nil {
			cleanupCopy := *repository.Cleanup
			cleanupCopy.PolicyNames = copyStrings(repository.Cleanup.PolicyNames)
			out.Repositories[i].Cleanup = &cleanupCopy
		}
	}
	for i, privilege := range in.Privileges {
		out.Privileges[i] = privilege
		out.Privileges[i].Actions = copyStrings(privilege.Actions)
	}
	for i, role := range in.Roles {
		out.Roles[i] = role
		out.Roles[i].Privileges = copyStrings(role.Privileges)
		out.Roles[i].Roles = copyStrings(role.Roles)
	}
	for i, permission := range in.UserRepositoryPermissions {
		out.UserRepositoryPermissions[i] = permission
		out.UserRepositoryPermissions[i].Privileges = copyStrings(permission.Privileges)
	}
	return out
}

// copyStrings clones a string slice.
func copyStrings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
