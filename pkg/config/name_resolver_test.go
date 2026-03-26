package config

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// TestResolveNamesWithSuffixSuffixMode verifies suffix mode rewrites all managed names and references.
func TestResolveNamesWithSuffixSuffixMode(t *testing.T) {
	const suffix = "20260324112233"

	cfg := &Config{
		NameMode: "suffix",
		Users: []User{
			{
				ID:           "repo-admin",
				EmailAddress: "repo-admin@example.com",
				Roles:        []string{"repository-manager", "nx-admin"},
			},
		},
		Repositories: []Repository{
			{Name: "maven-releases"},
		},
		Privileges: []Privilege{
			{
				Name:       "maven-deploy",
				Repository: "maven-releases",
				Actions:    []string{"READ"},
			},
		},
		Roles: []Role{
			{
				ID:         "repository-manager",
				Name:       "Repository Manager",
				Privileges: []string{"maven-deploy", "nx-repository-view-*-*-read"},
				Roles:      []string{"nx-admin"},
			},
		},
		UserRepositoryPermissions: []UserRepositoryPermission{
			{
				UserID:     "repo-admin",
				Repository: "maven-releases",
				Privileges: []string{"READ", "BROWSE"},
			},
		},
	}

	resolved, err := cfg.ResolveNamesWithSuffix(suffix)
	if err != nil {
		t.Fatalf("ResolveNamesWithSuffix() error = %v", err)
	}

	if got, want := resolved.Users[0].ID, "repo-admin-"+suffix; got != want {
		t.Fatalf("resolved user id = %q, want %q", got, want)
	}
	if got, want := resolved.Users[0].EmailAddress, "repo-admin+"+suffix+"@example.com"; got != want {
		t.Fatalf("resolved user email = %q, want %q", got, want)
	}
	if got, want := resolved.Users[0].Roles[0], "repository-manager-"+suffix; got != want {
		t.Fatalf("resolved user role ref = %q, want %q", got, want)
	}
	if got, want := resolved.Users[0].Roles[1], "nx-admin"; got != want {
		t.Fatalf("resolved built-in user role ref = %q, want %q", got, want)
	}

	if got, want := resolved.Repositories[0].Name, "maven-releases-"+suffix; got != want {
		t.Fatalf("resolved repository = %q, want %q", got, want)
	}

	if got, want := resolved.Privileges[0].Name, "maven-deploy-"+suffix; got != want {
		t.Fatalf("resolved privilege name = %q, want %q", got, want)
	}
	if got, want := resolved.Privileges[0].Repository, "maven-releases-"+suffix; got != want {
		t.Fatalf("resolved privilege repository = %q, want %q", got, want)
	}

	if got, want := resolved.Roles[0].ID, "repository-manager-"+suffix; got != want {
		t.Fatalf("resolved role id = %q, want %q", got, want)
	}
	if got, want := resolved.Roles[0].Name, "Repository Manager-"+suffix; got != want {
		t.Fatalf("resolved role name = %q, want %q", got, want)
	}
	if got, want := resolved.Roles[0].Privileges[0], "maven-deploy-"+suffix; got != want {
		t.Fatalf("resolved role privilege ref = %q, want %q", got, want)
	}
	if got, want := resolved.Roles[0].Privileges[1], "nx-repository-view-*-*-read"; got != want {
		t.Fatalf("resolved built-in role privilege ref = %q, want %q", got, want)
	}
	if got, want := resolved.Roles[0].Roles[0], "nx-admin"; got != want {
		t.Fatalf("resolved built-in role ref = %q, want %q", got, want)
	}

	if got, want := resolved.UserRepositoryPermissions[0].UserID, "repo-admin-"+suffix; got != want {
		t.Fatalf("resolved permission userId = %q, want %q", got, want)
	}
	if got, want := resolved.UserRepositoryPermissions[0].Repository, "maven-releases-"+suffix; got != want {
		t.Fatalf("resolved permission repository = %q, want %q", got, want)
	}

	if got, want := cfg.Users[0].ID, "repo-admin"; got != want {
		t.Fatalf("source config should stay unchanged, user id = %q, want %q", got, want)
	}
	if got, want := cfg.Repositories[0].Name, "maven-releases"; got != want {
		t.Fatalf("source config should stay unchanged, repo name = %q, want %q", got, want)
	}
}

func TestResolveNamesWithSuffixNoChangeByDefault(t *testing.T) {
	cfg := &Config{
		Users: []User{
			{ID: "repo-admin", EmailAddress: "repo-admin@example.com", Roles: []string{"repository-manager"}},
		},
		Repositories: []Repository{
			{Name: "maven-releases"},
		},
		Roles: []Role{
			{ID: "repository-manager", Name: "Repository Manager"},
		},
	}

	resolved, err := cfg.ResolveNamesWithSuffix("20260324112233")
	if err != nil {
		t.Fatalf("ResolveNamesWithSuffix() error = %v", err)
	}

	if got, want := resolved.Users[0].ID, "repo-admin"; got != want {
		t.Fatalf("resolved user id = %q, want %q", got, want)
	}
	if got, want := resolved.Repositories[0].Name, "maven-releases"; got != want {
		t.Fatalf("resolved repository = %q, want %q", got, want)
	}
	if got, want := resolved.Roles[0].ID, "repository-manager"; got != want {
		t.Fatalf("resolved role id = %q, want %q", got, want)
	}
}

// TestResolveNamesWithSuffixPreservesExplicitEmptyUserRoles ensures explicit empty roles keep [] instead of nil.
func TestResolveNamesWithSuffixPreservesExplicitEmptyUserRoles(t *testing.T) {
	cfg := &Config{
		Users: []User{
			{ID: "repo-admin", Roles: []string{}},
		},
	}

	resolved, err := cfg.ResolveNamesWithSuffix("20260324112233")
	if err != nil {
		t.Fatalf("ResolveNamesWithSuffix() error = %v", err)
	}

	if resolved.Users[0].Roles == nil {
		t.Fatal("resolved users[0].roles = nil, want explicit empty slice")
	}
	if len(resolved.Users[0].Roles) != 0 {
		t.Fatalf("resolved users[0].roles length = %d, want 0", len(resolved.Users[0].Roles))
	}
	if cfg.Users[0].Roles == nil {
		t.Fatal("source users[0].roles = nil, want explicit empty slice")
	}
}

func TestResolveNamesWithSuffixResourceOverride(t *testing.T) {
	const suffix = "20260324112233"

	cfg := &Config{
		NameMode: "name",
		Users: []User{
			{ID: "repo-admin", Roles: []string{"repository-manager"}},
		},
		Roles: []Role{
			{ID: "repository-manager", Name: "Repository Manager"},
		},
		Repositories: []Repository{
			{NameMode: "suffix", Name: "maven-releases"},
		},
		UserRepositoryPermissions: []UserRepositoryPermission{
			{UserID: "repo-admin", Repository: "maven-releases"},
		},
	}

	resolved, err := cfg.ResolveNamesWithSuffix(suffix)
	if err != nil {
		t.Fatalf("ResolveNamesWithSuffix() error = %v", err)
	}

	if got, want := resolved.Users[0].ID, "repo-admin"; got != want {
		t.Fatalf("resolved user id = %q, want %q", got, want)
	}
	if got, want := resolved.Repositories[0].Name, "maven-releases-"+suffix; got != want {
		t.Fatalf("resolved repository = %q, want %q", got, want)
	}
	if got, want := resolved.UserRepositoryPermissions[0].Repository, "maven-releases-"+suffix; got != want {
		t.Fatalf("resolved permission repository = %q, want %q", got, want)
	}
}

func TestResolveNamesInvalidMode(t *testing.T) {
	cfg := &Config{NameMode: "invalid"}
	if _, err := cfg.ResolveNamesWithSuffix("20260324112233"); err == nil {
		t.Fatal("ResolveNamesWithSuffix() expected error, got nil")
	}
}

// TestResolveNamesWithSuffixRequiresNonEmptySuffix verifies suffix mode rejects an empty suffix.
func TestResolveNamesWithSuffixRequiresNonEmptySuffix(t *testing.T) {
	cfg := &Config{
		NameMode: "suffix",
		Users: []User{
			{ID: "repo-admin"},
		},
	}

	_, err := cfg.ResolveNamesWithSuffix("")
	if err == nil {
		t.Fatal("ResolveNamesWithSuffix() expected error for empty suffix, got nil")
	}
	if !strings.Contains(err.Error(), "nameMode suffix requires a non-empty suffix") {
		t.Fatalf("ResolveNamesWithSuffix() error = %v, want contains %q", err, "nameMode suffix requires a non-empty suffix")
	}
}

// TestEmailWithSuffixWithoutAt verifies local fallback when email text has no '@'.
func TestEmailWithSuffixWithoutAt(t *testing.T) {
	const suffix = "20260324112233"

	got := emailWithSuffix("repo-admin", suffix)
	want := "repo-admin-" + suffix
	if got != want {
		t.Fatalf("emailWithSuffix() = %q, want %q", got, want)
	}
}

func TestResolveNamesGenerateSuffix(t *testing.T) {
	cfg := &Config{
		NameMode: "suffix",
		Users: []User{
			{ID: "repo-admin"},
		},
	}

	resolved, suffix, err := cfg.ResolveNames()
	if err != nil {
		t.Fatalf("ResolveNames() error = %v", err)
	}
	if suffix == "" {
		t.Fatal("ResolveNames() suffix should not be empty in suffix mode")
	}

	matched, err := regexp.MatchString(`^[0-9]{14}$`, suffix)
	if err != nil {
		t.Fatalf("regexp.MatchString() error = %v", err)
	}
	if !matched {
		t.Fatalf("ResolveNames() suffix format = %q, want 14-digit timestamp", suffix)
	}
	if !strings.HasSuffix(resolved.Users[0].ID, "-"+suffix) {
		t.Fatalf("resolved user id = %q should end with -%s", resolved.Users[0].ID, suffix)
	}
}

// TestCloneConfigDeepCopyMutableNestedFields ensures cloneConfig does not share mutable nested state.
func TestCloneConfigDeepCopyMutableNestedFields(t *testing.T) {
	cfg := &Config{
		Users: []User{
			{ID: "repo-admin", Roles: []string{"repository-manager"}},
		},
		Repositories: []Repository{
			{
				Name: "maven-releases",
				Proxy: &ProxyConfig{
					RemoteURL: "https://example.com",
					Authentication: &AuthConfig{
						Type:     "username",
						Username: "proxy-user",
						Password: "proxy-pass",
					},
				},
				Maven: &MavenConfig{
					VersionPolicy: "RELEASE",
					LayoutPolicy:  "STRICT",
				},
				Docker: &DockerConfig{
					HTTPPort:       18080,
					ForceBasicAuth: true,
				},
				Apt: &AptConfig{
					Distribution: "stable",
				},
				Cleanup: &CleanupConfig{
					PolicyNames: []string{"cleanup-old"},
				},
			},
		},
		Privileges: []Privilege{
			{Name: "maven-deploy", Actions: []string{"READ"}},
		},
		Roles: []Role{
			{ID: "repository-manager", Privileges: []string{"maven-deploy"}, Roles: []string{"nx-admin"}},
		},
		UserRepositoryPermissions: []UserRepositoryPermission{
			{UserID: "repo-admin", Repository: "maven-releases", Privileges: []string{"READ"}},
		},
	}

	cloned := cloneConfig(cfg)
	cloned.Users[0].Roles[0] = "changed-role"
	cloned.Repositories[0].Proxy.RemoteURL = "https://changed.example.com"
	cloned.Repositories[0].Proxy.Authentication.Username = "changed-proxy-user"
	cloned.Repositories[0].Maven.VersionPolicy = "MIXED"
	cloned.Repositories[0].Docker.HTTPPort = 28080
	cloned.Repositories[0].Apt.Distribution = "testing"
	cloned.Repositories[0].Cleanup.PolicyNames[0] = "changed-policy"
	cloned.Privileges[0].Actions[0] = "EDIT"
	cloned.Roles[0].Privileges[0] = "changed-privilege"
	cloned.Roles[0].Roles[0] = "changed-parent-role"
	cloned.UserRepositoryPermissions[0].Privileges[0] = "BROWSE"

	if cfg.Users[0].Roles[0] != "repository-manager" {
		t.Fatalf("source users roles mutated: got %q", cfg.Users[0].Roles[0])
	}
	if cfg.Repositories[0].Proxy.RemoteURL != "https://example.com" {
		t.Fatalf("source repository proxy remoteUrl mutated: got %q", cfg.Repositories[0].Proxy.RemoteURL)
	}
	if cfg.Repositories[0].Proxy.Authentication.Username != "proxy-user" {
		t.Fatalf("source repository proxy auth username mutated: got %q", cfg.Repositories[0].Proxy.Authentication.Username)
	}
	if cfg.Repositories[0].Maven.VersionPolicy != "RELEASE" {
		t.Fatalf("source repository maven versionPolicy mutated: got %q", cfg.Repositories[0].Maven.VersionPolicy)
	}
	if cfg.Repositories[0].Docker.HTTPPort != 18080 {
		t.Fatalf("source repository docker httpPort mutated: got %d", cfg.Repositories[0].Docker.HTTPPort)
	}
	if cfg.Repositories[0].Apt.Distribution != "stable" {
		t.Fatalf("source repository apt distribution mutated: got %q", cfg.Repositories[0].Apt.Distribution)
	}
	if cfg.Repositories[0].Cleanup.PolicyNames[0] != "cleanup-old" {
		t.Fatalf("source repository cleanup policyNames mutated: got %q", cfg.Repositories[0].Cleanup.PolicyNames[0])
	}
	if cfg.Privileges[0].Actions[0] != "READ" {
		t.Fatalf("source privilege actions mutated: got %q", cfg.Privileges[0].Actions[0])
	}
	if cfg.Roles[0].Privileges[0] != "maven-deploy" {
		t.Fatalf("source role privileges mutated: got %q", cfg.Roles[0].Privileges[0])
	}
	if cfg.Roles[0].Roles[0] != "nx-admin" {
		t.Fatalf("source role parent roles mutated: got %q", cfg.Roles[0].Roles[0])
	}
	if cfg.UserRepositoryPermissions[0].Privileges[0] != "READ" {
		t.Fatalf("source user repository permissions mutated: got %q", cfg.UserRepositoryPermissions[0].Privileges[0])
	}
}

// TestResolveNamesWithSuffixClearsNameModesForReuse ensures resolved config is idempotent for repeated create runs.
func TestResolveNamesWithSuffixClearsNameModesForReuse(t *testing.T) {
	const firstSuffix = "20260324112233"
	const secondSuffix = "20260324113344"

	cfg := &Config{
		NameMode: "suffix",
		Users: []User{
			{
				NameMode:     "suffix",
				ID:           "repo-admin",
				EmailAddress: "repo-admin@example.com",
				Roles:        []string{"repository-manager"},
			},
		},
		Repositories: []Repository{
			{NameMode: "suffix", Name: "maven-releases"},
		},
		Privileges: []Privilege{
			{
				NameMode:   "suffix",
				Name:       "maven-deploy",
				Repository: "maven-releases",
				Actions:    []string{"READ"},
			},
		},
		Roles: []Role{
			{
				NameMode:    "suffix",
				ID:          "repository-manager",
				Name:        "Repository Manager",
				Privileges:  []string{"maven-deploy"},
				Description: "Role used by repository admins",
			},
		},
		UserRepositoryPermissions: []UserRepositoryPermission{
			{
				UserID:     "repo-admin",
				Repository: "maven-releases",
				Privileges: []string{"READ"},
			},
		},
	}

	resolved, err := cfg.ResolveNamesWithSuffix(firstSuffix)
	if err != nil {
		t.Fatalf("ResolveNamesWithSuffix() first run error = %v", err)
	}

	if resolved.NameMode != "" {
		t.Fatalf("resolved config nameMode = %q, want empty", resolved.NameMode)
	}
	if resolved.Users[0].NameMode != "" {
		t.Fatalf("resolved users[0].nameMode = %q, want empty", resolved.Users[0].NameMode)
	}
	if resolved.Repositories[0].NameMode != "" {
		t.Fatalf("resolved repositories[0].nameMode = %q, want empty", resolved.Repositories[0].NameMode)
	}
	if resolved.Privileges[0].NameMode != "" {
		t.Fatalf("resolved privileges[0].nameMode = %q, want empty", resolved.Privileges[0].NameMode)
	}
	if resolved.Roles[0].NameMode != "" {
		t.Fatalf("resolved roles[0].nameMode = %q, want empty", resolved.Roles[0].NameMode)
	}

	rerun, err := resolved.ResolveNamesWithSuffix(secondSuffix)
	if err != nil {
		t.Fatalf("ResolveNamesWithSuffix() second run error = %v", err)
	}

	if !reflect.DeepEqual(rerun, resolved) {
		t.Fatalf("resolved config should be reusable without additional suffixing.\nfirst=%+v\nsecond=%+v", resolved, rerun)
	}
}
