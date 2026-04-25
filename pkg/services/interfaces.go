package services

import "github.com/tjg184/agent-smith/internal/models"

type InstallService interface {
	InstallSkill(repoURL, name string, opts InstallOptions) error
	InstallAgent(repoURL, name string, opts InstallOptions) error
	InstallCommand(repoURL, name string, opts InstallOptions) error
	InstallBulk(repoURL string, opts InstallOptions) error
}

type InstallOptions struct {
	Profile    string
	InstallDir string
}

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

type LinkOptions struct {
	TargetFilter string
	Profile      string
	AllProfiles  bool
	RepoURL      string
}

type UnlinkOptions struct {
	TargetFilter string
	Profile      string
	AllProfiles  bool
	Force        bool
	RepoURL      string
}

type LinkStatusOptions struct {
	AllProfiles   bool
	ProfileFilter []string
	LinkedOnly    bool
}

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

type ListProfileOptions struct {
	ProfileFilter []string
	ActiveOnly    bool
	TypeFilter    string
}

type MaterializeService interface {
	MaterializeComponent(componentType, name string, opts MaterializeOptions) error
	MaterializeByType(componentType string, opts MaterializeOptions) error
	MaterializeAll(opts MaterializeOptions) error
	ListMaterialized(opts ListMaterializedOptions) error
	ShowComponentInfo(componentType, name string, opts MaterializeInfoOptions) error
	ShowStatus(opts MaterializeStatusOptions) error
	UpdateMaterialized(opts MaterializeUpdateOptions) error
}

type MaterializeOptions struct {
	Target     string
	ProjectDir string
	Profile    string
	RepoURL    string
	Source     string
	Force      bool
	DryRun     bool
}

type ListMaterializedOptions struct {
	ProjectDir string
}

type MaterializeInfoOptions struct {
	Target     string
	ProjectDir string
	Source     string
}

type MaterializeStatusOptions struct {
	Target     string
	ProjectDir string
	DryRun     bool
}

type MaterializeUpdateOptions struct {
	Target     string
	ProjectDir string
	Profile    string
	Source     string
	Force      bool
	DryRun     bool
}

type UpdateService interface {
	UpdateComponent(componentType, name string, opts UpdateOptions) error
	UpdateAll(opts UpdateOptions) error
	CheckForUpdates(opts UpdateOptions) ([]UpdateInfo, error)
}

type UpdateOptions struct {
	Profile string
	RepoURL string
}

type UpdateInfo struct {
	Type      string
	Name      string
	Current   string
	Available string
}

type UninstallService interface {
	UninstallComponent(componentType, name string, opts UninstallOptions) error
	UninstallAllFromSource(repoURL string, opts UninstallOptions) error
}

type UninstallOptions struct {
	Profile string
	Source  string
	Force   bool
}

type TargetService interface {
	AddCustomTarget(name, path string) error
	RemoveCustomTarget(name string) error
	ListTargets() error
}

type StatusService interface {
	ShowSystemStatus() error
}

type FindService interface {
	FindSkills(query string, opts FindOptions) error
}

type FindOptions struct {
	Limit int
	JSON  bool
}

type ComponentLockService interface {
	LoadEntry(baseDir, componentType, componentName string) (*ComponentEntry, error)
	LoadEntryBySource(baseDir, componentType, componentName, sourceURL string) (*ComponentEntry, error)
	GetAllComponentNames(baseDir, componentType string) ([]string, error)
	FindComponentSources(baseDir, componentType, componentName string) ([]string, error)
	FindAllInstances(baseDir, componentType, componentName string) ([]*ComponentEntry, error)
	SaveEntry(baseDir, componentType, componentName string, entry *ComponentEntry) error
	RemoveEntry(baseDir, componentType, componentName string) error
	RemoveEntryBySource(baseDir, componentType, componentName, sourceURL string) error
	ResolveFilesystemName(baseDir, componentType, desiredName, sourceURL string) (string, error)
	HasConflict(baseDir, componentType, componentName string) (bool, error)
}

type ComponentEntry = models.ComponentEntry
