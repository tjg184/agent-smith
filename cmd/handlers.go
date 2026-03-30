package cmd

// InstallHandlers groups handler functions for install commands.
type InstallHandlers struct {
	AddSkill   func(repoURL, name, profile, targetDir string, global bool)
	AddAgent   func(repoURL, name, profile, targetDir string, global bool)
	AddCommand func(repoURL, name, profile, targetDir string, global bool)
	AddAll     func(repoURL, profile, targetDir string, global bool)
}

// UpdateHandlers groups handler functions for update commands.
type UpdateHandlers struct {
	Update    func(componentType, componentName, profile string)
	UpdateAll func(profile string)
}

// LinkHandlers groups handler functions for link commands.
type LinkHandlers struct {
	Link       func(componentType, componentName, targetFilter, profile string)
	LinkAll    func(targetFilter, profile string, allProfiles bool)
	LinkType   func(componentType, targetFilter, profile string)
	AutoLink   func()
	ListLinks  func()
	LinkStatus func(allProfiles bool, profileFilter []string, linkedOnly bool)
}

// UnlinkHandlers groups handler functions for unlink commands.
type UnlinkHandlers struct {
	Unlink                func(componentType, componentName, targetFilter string)
	UnlinkWithProfile     func(componentType, componentName, targetFilter, profile string)
	UnlinkAll             func(targetFilter string, force bool, allProfiles bool)
	UnlinkAllWithProfile  func(targetFilter string, force bool, allProfiles bool, profile string)
	UnlinkType            func(componentType, targetFilter string, force bool)
	UnlinkTypeWithProfile func(componentType, targetFilter string, force bool, profile string)
}

// UninstallHandlers groups handler functions for uninstall commands.
type UninstallHandlers struct {
	Uninstall    func(componentType, componentName, profile, source string)
	UninstallAll func(repoURL string, force bool)
}

// ProfileHandlers groups handler functions for profile commands.
type ProfileHandlers struct {
	List       func(profileFilter []string, activeOnly bool, typeFilter string)
	Show       func(profileName string)
	Create     func(profileName string)
	Delete     func(profileName string)
	Activate   func(profileName string)
	Deactivate func()
	Add        func(componentType, profileName, componentName string)
	Copy       func(componentType, sourceProfile, targetProfile, componentName string)
	Remove     func(componentType, profileName, componentName string)
	CherryPick func(targetProfile string, sourceProfiles []string)
	Share      func(profileName, outputFile string)
	Rename     func(oldName, newName string)
}

// StatusHandlers groups handler functions for status commands.
type StatusHandlers struct {
	Status func()
}

// TargetHandlers groups handler functions for target commands.
type TargetHandlers struct {
	Add    func(name, path string)
	Remove func(name string)
	List   func()
}

// MaterializeHandlers groups handler functions for materialize commands.
type MaterializeHandlers struct {
	Component func(componentType, componentName, target, projectDir string, force, dryRun bool, fromProfile, source string)
	Type      func(componentType, target, projectDir string, force, dryRun bool, fromProfile string)
	All       func(target, projectDir string, force, dryRun bool, fromProfile string)
	List      func(projectDir string)
	Info      func(componentType, componentName, target, projectDir, source string)
	Status    func(target, projectDir string)
	Update    func(target, projectDir, source string, force, dryRun bool)
}

// FindHandlers groups handler functions for find commands.
type FindHandlers struct {
	FindSkill func(query string, limit int, jsonOutput bool)
}

// Handlers is the root struct grouping all command handler functions by domain.
// Construct it in the container and pass it to Register.
type Handlers struct {
	Install     InstallHandlers
	Update      UpdateHandlers
	Link        LinkHandlers
	Unlink      UnlinkHandlers
	Uninstall   UninstallHandlers
	Profile     ProfileHandlers
	Status      StatusHandlers
	Target      TargetHandlers
	Materialize MaterializeHandlers
	Find        FindHandlers
}

// Package-level vars referenced by all cmd/*.go files — do not rename.
var (
	handleAddSkill              func(repoURL, name, profile, targetDir string, global bool)
	handleAddAgent              func(repoURL, name, profile, targetDir string, global bool)
	handleAddCommand            func(repoURL, name, profile, targetDir string, global bool)
	handleAddAll                func(repoURL, profile, targetDir string, global bool)
	handleUpdate                func(componentType, componentName, profile string)
	handleUpdateAll             func(profile string)
	handleLink                  func(componentType, componentName, targetFilter, profile string)
	handleLinkAll               func(targetFilter, profile string, allProfiles bool)
	handleLinkType              func(componentType, targetFilter, profile string)
	handleAutoLink              func()
	handleListLinks             func()
	handleLinkStatus            func(allProfiles bool, profileFilter []string, linkedOnly bool)
	handleUnlink                func(componentType, componentName, targetFilter string)
	handleUnlinkWithProfile     func(componentType, componentName, targetFilter, profile string)
	handleUnlinkAll             func(targetFilter string, force bool, allProfiles bool)
	handleUnlinkAllWithProfile  func(targetFilter string, force bool, allProfiles bool, profile string)
	handleUnlinkType            func(componentType, targetFilter string, force bool)
	handleUnlinkTypeWithProfile func(componentType, targetFilter string, force bool, profile string)
	handleUninstall             func(componentType, componentName, profile, source string)
	handleUninstallAll          func(repoURL string, force bool)
	handleProfilesList          func(profileFilter []string, activeOnly bool, typeFilter string)
	handleProfilesShow          func(profileName string)
	handleProfilesCreate        func(profileName string)
	handleProfilesDelete        func(profileName string)
	handleProfilesActivate      func(profileName string)
	handleProfilesDeactivate    func()
	handleProfilesAdd           func(componentType, profileName, componentName string)
	handleProfilesCopy          func(componentType, sourceProfile, targetProfile, componentName string)
	handleProfilesRemove        func(componentType, profileName, componentName string)
	handleProfilesCherryPick    func(targetProfile string, sourceProfiles []string)
	handleProfilesShare         func(profileName, outputFile string)
	handleProfilesRename        func(oldName, newName string)
	handleStatus                func()
	handleTargetAdd             func(name, path string)
	handleTargetRemove          func(name string)
	handleTargetList            func()
	handleMaterializeComponent  func(componentType, componentName, target, projectDir string, force, dryRun bool, fromProfile, source string)
	handleMaterializeType       func(componentType, target, projectDir string, force, dryRun bool, fromProfile string)
	handleMaterializeAll        func(target, projectDir string, force, dryRun bool, fromProfile string)
	handleMaterializeList       func(projectDir string)
	handleMaterializeInfo       func(componentType, componentName, target, projectDir, source string)
	handleMaterializeStatus     func(target, projectDir string)
	handleMaterializeUpdate     func(target, projectDir, source string, force, dryRun bool)
	handleFindSkill             func(query string, limit int, jsonOutput bool)
)

// Register assigns all package-level handler vars from h.
func Register(h *Handlers) {
	handleAddSkill = h.Install.AddSkill
	handleAddAgent = h.Install.AddAgent
	handleAddCommand = h.Install.AddCommand
	handleAddAll = h.Install.AddAll
	handleUpdate = h.Update.Update
	handleUpdateAll = h.Update.UpdateAll
	handleLink = h.Link.Link
	handleLinkAll = h.Link.LinkAll
	handleLinkType = h.Link.LinkType
	handleAutoLink = h.Link.AutoLink
	handleListLinks = h.Link.ListLinks
	handleLinkStatus = h.Link.LinkStatus
	handleUnlink = h.Unlink.Unlink
	handleUnlinkWithProfile = h.Unlink.UnlinkWithProfile
	handleUnlinkAll = h.Unlink.UnlinkAll
	handleUnlinkAllWithProfile = h.Unlink.UnlinkAllWithProfile
	handleUnlinkType = h.Unlink.UnlinkType
	handleUnlinkTypeWithProfile = h.Unlink.UnlinkTypeWithProfile
	handleUninstall = h.Uninstall.Uninstall
	handleUninstallAll = h.Uninstall.UninstallAll
	handleProfilesList = h.Profile.List
	handleProfilesShow = h.Profile.Show
	handleProfilesCreate = h.Profile.Create
	handleProfilesDelete = h.Profile.Delete
	handleProfilesActivate = h.Profile.Activate
	handleProfilesDeactivate = h.Profile.Deactivate
	handleProfilesAdd = h.Profile.Add
	handleProfilesCopy = h.Profile.Copy
	handleProfilesRemove = h.Profile.Remove
	handleProfilesCherryPick = h.Profile.CherryPick
	handleProfilesShare = h.Profile.Share
	handleProfilesRename = h.Profile.Rename
	handleStatus = h.Status.Status
	handleTargetAdd = h.Target.Add
	handleTargetRemove = h.Target.Remove
	handleTargetList = h.Target.List
	handleMaterializeComponent = h.Materialize.Component
	handleMaterializeType = h.Materialize.Type
	handleMaterializeAll = h.Materialize.All
	handleMaterializeList = h.Materialize.List
	handleMaterializeInfo = h.Materialize.Info
	handleMaterializeStatus = h.Materialize.Status
	handleMaterializeUpdate = h.Materialize.Update
	handleFindSkill = h.Find.FindSkill
}
