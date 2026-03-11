package cmd

// These functions will be implemented in main.go to keep existing logic
var (
	handleAddSkill              func(repoURL, name, profile, targetDir string)
	handleAddAgent              func(repoURL, name, profile, targetDir string)
	handleAddCommand            func(repoURL, name, profile, targetDir string)
	handleAddAll                func(repoURL, profile, targetDir string)
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
	handleUninstall             func(componentType, componentName, profile string)
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

// SetHandlers assigns the handler functions that implement the command logic
func SetHandlers(
	addSkill func(repoURL, name, profile, targetDir string),
	addAgent func(repoURL, name, profile, targetDir string),
	addCommand func(repoURL, name, profile, targetDir string),
	addAll func(repoURL, profile, targetDir string),
	update func(componentType, componentName, profile string),
	updateAll func(profile string),
	link func(componentType, componentName, targetFilter, profile string),
	linkAll func(targetFilter, profile string, allProfiles bool),
	linkType func(componentType, targetFilter, profile string),
	autoLink func(),
	listLinks func(),
	linkStatus func(allProfiles bool, profileFilter []string, linkedOnly bool),
	unlink func(componentType, componentName, targetFilter string),
	unlinkWithProfile func(componentType, componentName, targetFilter, profile string),
	unlinkAll func(targetFilter string, force bool, allProfiles bool),
	unlinkAllWithProfile func(targetFilter string, force bool, allProfiles bool, profile string),
	unlinkType func(componentType, targetFilter string, force bool),
	unlinkTypeWithProfile func(componentType, targetFilter string, force bool, profile string),
	uninstall func(componentType, componentName, profile string),
	uninstallAll func(repoURL string, force bool),
	profilesList func(profileFilter []string, activeOnly bool, typeFilter string),
	profilesShow func(profileName string),
	profilesCreate func(profileName string),
	profilesDelete func(profileName string),
	profilesActivate func(profileName string),
	profilesDeactivate func(),
	profilesAdd func(componentType, profileName, componentName string),
	profilesCopy func(componentType, sourceProfile, targetProfile, componentName string),
	profilesRemove func(componentType, profileName, componentName string),
	profilesCherryPick func(targetProfile string, sourceProfiles []string),
	profilesShare func(profileName, outputFile string),
	profilesRename func(oldName, newName string),
	status func(),
	targetAdd func(name, path string),
	targetRemove func(name string),
	targetList func(),
	materializeComponent func(componentType, componentName, target, projectDir string, force, dryRun bool, fromProfile, source string),
	materializeType func(componentType, target, projectDir string, force, dryRun bool, fromProfile string),
	materializeAll func(target, projectDir string, force, dryRun bool, fromProfile string),
	materializeList func(projectDir string),
	materializeInfo func(componentType, componentName, target, projectDir, source string),
	materializeStatus func(target, projectDir string),
	materializeUpdate func(target, projectDir, source string, force, dryRun bool),
	findSkill func(query string, limit int, jsonOutput bool),
) {
	handleAddSkill = addSkill
	handleAddAgent = addAgent
	handleAddCommand = addCommand
	handleAddAll = addAll
	handleUpdate = update
	handleUpdateAll = updateAll
	handleLink = link
	handleLinkAll = linkAll
	handleLinkType = linkType
	handleAutoLink = autoLink
	handleListLinks = listLinks
	handleLinkStatus = linkStatus
	handleUnlink = unlink
	handleUnlinkWithProfile = unlinkWithProfile
	handleUnlinkAll = unlinkAll
	handleUnlinkAllWithProfile = unlinkAllWithProfile
	handleUnlinkType = unlinkType
	handleUnlinkTypeWithProfile = unlinkTypeWithProfile
	handleUninstall = uninstall
	handleUninstallAll = uninstallAll
	handleProfilesList = profilesList
	handleProfilesShow = profilesShow
	handleProfilesCreate = profilesCreate
	handleProfilesDelete = profilesDelete
	handleProfilesActivate = profilesActivate
	handleProfilesDeactivate = profilesDeactivate
	handleProfilesAdd = profilesAdd
	handleProfilesCopy = profilesCopy
	handleProfilesRemove = profilesRemove
	handleProfilesCherryPick = profilesCherryPick
	handleProfilesShare = profilesShare
	handleProfilesRename = profilesRename
	handleStatus = status
	handleTargetAdd = targetAdd
	handleTargetRemove = targetRemove
	handleTargetList = targetList
	handleMaterializeComponent = materializeComponent
	handleMaterializeType = materializeType
	handleMaterializeAll = materializeAll
	handleMaterializeList = materializeList
	handleMaterializeInfo = materializeInfo
	handleMaterializeStatus = materializeStatus
	handleMaterializeUpdate = materializeUpdate
	handleFindSkill = findSkill
}
