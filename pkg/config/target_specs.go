package config

type targetSpec struct {
	name        string
	displayName string
	globalDir   string // raw path, may contain ~
	projectDir  string
	isUniversal bool
}

// builtInTargetSpecs is the single source of truth for all built-in agents.
// Add a new agent by appending one line here — no other files need changing.
// Order determines detection priority.
var builtInTargetSpecs = []targetSpec{
	{"opencode", "OpenCode", "~/.config/opencode", ".opencode", false},
	{"claudecode", "Claude Code", "~/.claude", ".claude", false},
	{"copilot", "GitHub Copilot", "~/.copilot", ".github", false},
	{"universal", "Universal", "~/.agents", ".agents", true},
}
