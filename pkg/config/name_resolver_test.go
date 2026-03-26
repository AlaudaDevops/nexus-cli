package config

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestResolveNamesWithSuffixPrefixMode(t *testing.T) {
	const suffix = "20260324112233"

	cfg := &Config{
		NameMode: "prefix",
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
			{NameMode: "prefix", Name: "maven-releases"},
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

func TestResolveNamesGenerateSuffix(t *testing.T) {
	cfg := &Config{
		NameMode: "prefix",
		Users: []User{
			{ID: "repo-admin"},
		},
	}

	resolved, suffix, err := cfg.ResolveNames()
	if err != nil {
		t.Fatalf("ResolveNames() error = %v", err)
	}
	if suffix == "" {
		t.Fatal("ResolveNames() suffix should not be empty in prefix mode")
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

// TestResolveNamesWithSuffixClearsNameModesForReuse ensures resolved config is idempotent for repeated create runs.
func TestResolveNamesWithSuffixClearsNameModesForReuse(t *testing.T) {
	const firstSuffix = "20260324112233"
	const secondSuffix = "20260324113344"

	cfg := &Config{
		NameMode: "prefix",
		Users: []User{
			{
				NameMode:     "prefix",
				ID:           "repo-admin",
				EmailAddress: "repo-admin@example.com",
				Roles:        []string{"repository-manager"},
			},
		},
		Repositories: []Repository{
			{NameMode: "prefix", Name: "maven-releases"},
		},
		Privileges: []Privilege{
			{
				NameMode:   "prefix",
				Name:       "maven-deploy",
				Repository: "maven-releases",
				Actions:    []string{"READ"},
			},
		},
		Roles: []Role{
			{
				NameMode:    "prefix",
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
