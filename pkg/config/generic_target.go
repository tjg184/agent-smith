package config

type genericTarget struct {
	baseTarget
	name        string
	displayName string
	isUniversal bool
}

func (t *genericTarget) GetName() string {
	return t.name
}

func (t *genericTarget) GetDisplayName() string {
	return t.displayName
}

func (t *genericTarget) IsUniversalTarget() bool {
	return t.isUniversal
}
