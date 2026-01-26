package profiles

import (
	"testing"
)

func TestProfile_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		profile *Profile
		want    bool
	}{
		{
			name: "valid profile with agents only",
			profile: &Profile{
				Name:      "test",
				BasePath:  "/home/user/.agents/profiles/test",
				HasAgents: true,
			},
			want: true,
		},
		{
			name: "valid profile with skills only",
			profile: &Profile{
				Name:      "test",
				BasePath:  "/home/user/.agents/profiles/test",
				HasSkills: true,
			},
			want: true,
		},
		{
			name: "valid profile with commands only",
			profile: &Profile{
				Name:        "test",
				BasePath:    "/home/user/.agents/profiles/test",
				HasCommands: true,
			},
			want: true,
		},
		{
			name: "valid profile with all components",
			profile: &Profile{
				Name:        "test",
				BasePath:    "/home/user/.agents/profiles/test",
				HasAgents:   true,
				HasSkills:   true,
				HasCommands: true,
			},
			want: true,
		},
		{
			name: "invalid profile with no components",
			profile: &Profile{
				Name:     "test",
				BasePath: "/home/user/.agents/profiles/test",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.profile.IsValid(); got != tt.want {
				t.Errorf("Profile.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfile_GetAgentsDir(t *testing.T) {
	profile := &Profile{
		Name:      "test",
		BasePath:  "/home/user/.agents/profiles/test",
		HasAgents: true,
	}

	expected := "/home/user/.agents/profiles/test/agents"
	if got := profile.GetAgentsDir(); got != expected {
		t.Errorf("Profile.GetAgentsDir() = %v, want %v", got, expected)
	}
}

func TestProfile_GetSkillsDir(t *testing.T) {
	profile := &Profile{
		Name:      "test",
		BasePath:  "/home/user/.agents/profiles/test",
		HasSkills: true,
	}

	expected := "/home/user/.agents/profiles/test/skills"
	if got := profile.GetSkillsDir(); got != expected {
		t.Errorf("Profile.GetSkillsDir() = %v, want %v", got, expected)
	}
}

func TestProfile_GetCommandsDir(t *testing.T) {
	profile := &Profile{
		Name:        "test",
		BasePath:    "/home/user/.agents/profiles/test",
		HasCommands: true,
	}

	expected := "/home/user/.agents/profiles/test/commands"
	if got := profile.GetCommandsDir(); got != expected {
		t.Errorf("Profile.GetCommandsDir() = %v, want %v", got, expected)
	}
}
