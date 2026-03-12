package services

import "github.com/tjg184/agent-smith/internal/models"

// InstallService handles installation of components (skills, agents, commands)
type InstallService interface {
	InstallSkill(repoURL, name string, opts InstallOptions) error
	InstallAgent(repoURL, name string, opts InstallOptions) error
	InstallCommand(repoURL, name string, opts InstallOptions) error
	InstallBulk(repoURL string, opts InstallOptions) error
}

// InstallOptions configures component installation
type InstallOptions struct {
	Profile    string // Profile to install to (empty = base directory)
	InstallDir string // Custom installation directory (empty = use default)
}

// LinkService handles linking/unlinking components to targets
type LinkService interface {
	LinkComponent(componentType, name string, opts LinkOptions) error
	LinkAll(opts LinkOptions) error
	LinkByType(componentType string, opts LinkOptions) error
	UnlinkComponent(componentType, name string, opts UnlinkOptions) error
	UnlinkAll(opts UnlinkOptions) error
	UnlinkByType(componentType string, opts UnlinkOptions) error
	AutoLinkRepositories() error
	ListLinked() error
	ShowStatus(opts LinkStatusOptions) error
}

// LinkOptions configures component linking
type LinkOptions struct {
	TargetFilter string // Filter for specific target (e.g., "opencode", "claudecode")
	Profile      string // Explicit profile to use (bypasses active profile)
	AllProfiles  bool   // Link from all profiles
}

// UnlinkOptions configures component unlinking
type UnlinkOptions struct {
	TargetFilter string // Filter for specific target
	Profile      string // Explicit profile to use
	AllProfiles  bool   // Unlink from all profiles
	Force        bool   // Force unlink without confirmation
}

// LinkStatusOptions configures link status display
type LinkStatusOptions struct {
	AllProfiles   bool     // Show status for all profiles
	ProfileFilter []string // Filter by specific profile names
	LinkedOnly    bool     // Show only components with at least one link
}

// ProfileService handles profile management operations
type ProfileService interface {
	ListProfiles(opts ListProfileOptions) error
	ShowProfile(name string) error
	CreateProfile(name string) error
	DeleteProfile(name string) error
	ActivateProfile(name string) error
	DeactivateProfile() error
	AddComponent(componentType, profileName, componentName string) error
	CopyComponent(sourceProfile, targetProfile, componentType, componentName string) error
	RemoveComponent(profileName, componentType, componentName string) error
	CherryPickComponents(targetProfile string, sourceProfiles []string) error
	ShareProfile(name, outputPath string) error
	RenameProfile(oldName, newName string) error
}

// ListProfileOptions configures profile listing
type ListProfileOptions struct {
	ProfileFilter []string // Filter by specific profile names
	ActiveOnly    bool     // Show only active profile
	TypeFilter    string   // Filter by type: "repo" or "user"
}

// MaterializeService handles materializing components to projects
type MaterializeService interface {
	MaterializeComponent(componentType, name string, opts MaterializeOptions) error
	MaterializeByType(componentType string, opts MaterializeOptions) error
	MaterializeAll(opts MaterializeOptions) error
	ListMaterialized(opts ListMaterializedOptions) error
	ShowComponentInfo(componentType, name string, opts MaterializeInfoOptions) error
	ShowStatus(opts MaterializeStatusOptions) error
	UpdateMaterialized(opts MaterializeUpdateOptions) error
}

// MaterializeOptions configures component materialization
type MaterializeOptions struct {
	Target     string // Target to materialize to (e.g., "opencode", "claudecode", "all")
	ProjectDir string // Project directory (empty = auto-detect)
	Profile    string // Profile to materialize from (empty = active profile or base)
	Source     string // Source URL to materialize from (for disambiguation when component exists in multiple sources)
	Force      bool   // Force overwrite existing components
	DryRun     bool   // Simulate without making changes
}

// ListMaterializedOptions configures listing materialized components
type ListMaterializedOptions struct {
	ProjectDir string // Project directory (empty = auto-detect)
}

// MaterializeInfoOptions configures component info display
type MaterializeInfoOptions struct {
	Target     string // Target to show info for
	ProjectDir string // Project directory (empty = auto-detect)
	Source     string // Source URL filter (for disambiguation)
}

// MaterializeStatusOptions configures materialization status display
type MaterializeStatusOptions struct {
	Target     string // Target to check (empty = all targets)
	ProjectDir string // Project directory (empty = auto-detect)
	DryRun     bool   // Simulate status check
}

// MaterializeUpdateOptions configures materialization updates
type MaterializeUpdateOptions struct {
	Target     string // Target to update (empty = all targets)
	ProjectDir string // Project directory (empty = auto-detect)
	Profile    string // Profile to source components from (empty = active profile or base)
	Source     string // Source URL filter (for disambiguation)
	Force      bool   // Force re-materialization even if in sync
	DryRun     bool   // Simulate updates
}

// UpdateService handles updating components
type UpdateService interface {
	UpdateComponent(componentType, name string, opts UpdateOptions) error
	UpdateAll(opts UpdateOptions) error
	CheckForUpdates(opts UpdateOptions) ([]UpdateInfo, error)
}

// UpdateOptions configures component updates
type UpdateOptions struct {
	Profile string // Profile to update (empty = active profile or base)
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	Type      string // Component type
	Name      string // Component name
	Current   string // Current version/commit
	Available string // Available version/commit
}

// UninstallService handles uninstalling components
type UninstallService interface {
	UninstallComponent(componentType, name string, opts UninstallOptions) error
	UninstallAllFromSource(repoURL string, opts UninstallOptions) error
}

// UninstallOptions configures component uninstallation
type UninstallOptions struct {
	Profile string // Profile to uninstall from (empty = base directory)
	Source  string // Source URL filter (for disambiguation when component exists in multiple sources)
	Force   bool   // Force uninstall without confirmation
}

// TargetService handles custom target management
type TargetService interface {
	AddCustomTarget(name, path string) error
	RemoveCustomTarget(name string) error
	ListTargets() error
}

// StatusService handles system status display
type StatusService interface {
	ShowSystemStatus() error
}

// FindService handles searching for components in remote registries
type FindService interface {
	FindSkills(query string, opts FindOptions) error
}

// FindOptions configures component search
type FindOptions struct {
	Limit int  // Max results to display (default: 20)
	JSON  bool // Output as JSON for scripting
}

// ComponentLockService handles reading and writing component lock files
type ComponentLockService interface {
	// Read operations
	LoadEntry(baseDir, componentType, componentName string) (*ComponentEntry, error)
	LoadEntryBySource(baseDir, componentType, componentName, sourceURL string) (*ComponentEntry, error)
	GetAllComponentNames(baseDir, componentType string) ([]string, error)
	FindComponentSources(baseDir, componentType, componentName string) ([]string, error)
	FindAllInstances(baseDir, componentType, componentName string) ([]*ComponentEntry, error)

	// Write operations
	SaveEntry(baseDir, componentType, componentName string, entry *ComponentEntry) error
	RemoveEntry(baseDir, componentType, componentName string) error
	RemoveEntryBySource(baseDir, componentType, componentName, sourceURL string) error

	// Utility operations
	ResolveFilesystemName(baseDir, componentType, desiredName, sourceURL string) (string, error)
	HasConflict(baseDir, componentType, componentName string) (bool, error)
}

// ComponentEntry is re-exported from models for service interface
type ComponentEntry = models.ComponentEntry
