package services

import "testing"

func TestResolveClaimPath(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		path string
		want interface{}
	}{
		{
			name: "flat key",
			data: map[string]interface{}{"role": "admin"},
			path: "role",
			want: "admin",
		},
		{
			name: "nested dot notation",
			data: map[string]interface{}{
				"realm_access": map[string]interface{}{
					"roles": []interface{}{"traceway-admin", "offline_access"},
				},
			},
			path: "realm_access.roles",
			want: []interface{}{"traceway-admin", "offline_access"},
		},
		{
			name: "missing intermediate key",
			data: map[string]interface{}{"a": map[string]interface{}{}},
			path: "a.b.c",
			want: nil,
		},
		{
			name: "non-map intermediate",
			data: map[string]interface{}{"a": "string"},
			path: "a.b",
			want: nil,
		},
		{
			name: "missing top-level key",
			data: map[string]interface{}{},
			path: "groups",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveClaimPath(tt.data, tt.path)
			if !deepEqual(got, tt.want) {
				t.Errorf("resolveClaimPath(%v, %q) = %v, want %v", tt.data, tt.path, got, tt.want)
			}
		})
	}
}

func TestNormalizeClaimToStrings(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want []string
	}{
		{
			name: "string value",
			v:    "admin",
			want: []string{"admin"},
		},
		{
			name: "string slice",
			v:    []interface{}{"admin", "user"},
			want: []string{"admin", "user"},
		},
		{
			name: "mixed slice drops non-strings",
			v:    []interface{}{"admin", 42, nil, "readonly"},
			want: []string{"admin", "readonly"},
		},
		{
			name: "empty slice",
			v:    []interface{}{},
			want: []string{},
		},
		{
			name: "nil returns nil",
			v:    nil,
			want: nil,
		},
		{
			name: "unsupported type returns nil",
			v:    map[string]interface{}{"x": "y"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeClaimToStrings(tt.v)
			if len(got) != len(tt.want) {
				t.Errorf("normalizeClaimToStrings(%v) = %v, want %v", tt.v, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("normalizeClaimToStrings(%v)[%d] = %q, want %q", tt.v, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestResolveRole(t *testing.T) {
	svc := &oauthService{
		oidcRoleClaim: "realm_access.roles",
		oidcRoleMap: map[string]string{
			"kc-admin":    "admin",
			"kc-user":     "user",
			"kc-readonly": "readonly",
		},
	}

	tests := []struct {
		name    string
		rawData map[string]interface{}
		want    string
	}{
		{
			name: "maps single matching role",
			rawData: map[string]interface{}{
				"realm_access": map[string]interface{}{
					"roles": []interface{}{"kc-admin"},
				},
			},
			want: "admin",
		},
		{
			name: "highest priority wins",
			rawData: map[string]interface{}{
				"realm_access": map[string]interface{}{
					"roles": []interface{}{"kc-readonly", "kc-admin", "kc-user"},
				},
			},
			want: "admin",
		},
		{
			name: "no matching roles returns empty string",
			rawData: map[string]interface{}{
				"realm_access": map[string]interface{}{
					"roles": []interface{}{"some-other-role"},
				},
			},
			want: "",
		},
		{
			name:    "missing claim returns empty string",
			rawData: map[string]interface{}{},
			want:    "",
		},
		{
			name: "maps readonly",
			rawData: map[string]interface{}{
				"realm_access": map[string]interface{}{
					"roles": []interface{}{"kc-readonly"},
				},
			},
			want: "readonly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.ResolveRole(tt.rawData)
			if got != tt.want {
				t.Errorf("ResolveRole() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveRoleFlatClaim(t *testing.T) {
	svc := &oauthService{
		oidcRoleClaim: "groups",
		oidcRoleMap:   map[string]string{"/admins": "admin", "/users": "user"},
	}

	rawData := map[string]interface{}{
		"groups": []interface{}{"/admins", "/users"},
	}
	got := svc.ResolveRole(rawData)
	if got != "admin" {
		t.Errorf("ResolveRole() = %q, want %q", got, "admin")
	}
}

func TestExtractURLOrigin(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://keycloak.example.com/realms/myrealm/protocol/openid-connect/auth", "https://keycloak.example.com"},
		{"http://localhost:8080/auth", "http://localhost:8080"},
		{"https://example.com", "https://example.com"},
		{"not-a-url", "not-a-url"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractURLOrigin(tt.input)
			if got != tt.want {
				t.Errorf("extractURLOrigin(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	aSlice, aIsSlice := a.([]interface{})
	bSlice, bIsSlice := b.([]interface{})
	if aIsSlice && bIsSlice {
		if len(aSlice) != len(bSlice) {
			return false
		}
		for i := range aSlice {
			if !deepEqual(aSlice[i], bSlice[i]) {
				return false
			}
		}
		return true
	}
	return a == b
}
