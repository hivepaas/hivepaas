package entity

import "time"

type TaskSystemCleanupOutput struct {
	DBCleanup      *DBCleanupOutput      `json:"dbCleanup"`
	ClusterCleanup *ClusterCleanupOutput `json:"clusterCleanup"`
	CacheCleanup   *CacheCleanupOutput   `json:"cacheCleanup"`
	FileCleanup    *FileCleanupOutput    `json:"fileCleanup"`
	// SystemApps is nil when the system apps were not synced.
	SystemApps *SystemAppsSyncOutput `json:"systemApps,omitempty"`
}

// SystemAppsSyncOutput is what syncing the system apps found and did.
type SystemAppsSyncOutput struct {
	// Skipped says why nothing was synced: a system update running.
	Skipped string                 `json:"skipped,omitempty"`
	Apps    []*SystemAppSyncOutput `json:"apps,omitempty"`
	OBI     *OBISyncOutput         `json:"obi,omitempty"`
}

// SystemAppSyncAction is what syncing did to a system app.
type SystemAppSyncAction string

const (
	// SystemAppSyncNone: the app is as its settings say, or none is wanted.
	SystemAppSyncNone SystemAppSyncAction = "none"
	// SystemAppSyncProvisioned: the settings want the app, and it was missing.
	SystemAppSyncProvisioned SystemAppSyncAction = "provisioned"
	// SystemAppSyncUpdated: the app was deployed with what its settings and
	// the release say - its image, its command.
	SystemAppSyncUpdated SystemAppSyncAction = "updated"
	// SystemAppSyncRecreated: the app's service was gone, and the app was
	// removed and provisioned again, its data kept.
	SystemAppSyncRecreated SystemAppSyncAction = "recreated"
	// SystemAppSyncRemoved: the settings no longer want the app; its data was
	// kept.
	SystemAppSyncRemoved SystemAppSyncAction = "removed"
	// SystemAppSyncReported: the app runs short of what it should, and that is
	// left to a person: Problem says how.
	SystemAppSyncReported SystemAppSyncAction = "reported"
	// SystemAppSyncSkipped: the app is being deployed or updated; the next run
	// looks again.
	SystemAppSyncSkipped SystemAppSyncAction = "skipped"
	// SystemAppSyncFailed: syncing it failed: Error says why.
	SystemAppSyncFailed SystemAppSyncAction = "failed"
)

// SystemAppSyncOutput is one system app's sync.
type SystemAppSyncOutput struct {
	Key   string `json:"key"`
	Name  string `json:"name,omitempty"`
	AppID string `json:"appId,omitempty"`
	// PreviousAppID is the app removed for one recreated, which has AppID.
	PreviousAppID string              `json:"previousAppId,omitempty"`
	Action        SystemAppSyncAction `json:"action"`
	Problem       string              `json:"problem,omitempty"`
	Error         string              `json:"error,omitempty"`
}

// OBISyncAction is what checking OBI did about a node.
type OBISyncAction string

const (
	OBISyncNone OBISyncAction = "none"
	// OBISyncNodeRemoved: the node is no longer in the cluster, and was taken
	// off the nodes that run OBI.
	OBISyncNodeRemoved OBISyncAction = "node-removed"
	// OBISyncCacheCleared: the node's agent did not do what the settings say,
	// and the agents were made to read them again.
	OBISyncCacheCleared OBISyncAction = "cache-cleared"
	// OBISyncReported: the node does not run OBI as it should, and that is left
	// to a person: Problem says why.
	OBISyncReported OBISyncAction = "reported"
)

// OBISyncOutput is the check of OBI on the nodes.
type OBISyncOutput struct {
	// Unknown says why the nodes could not be checked: their status rows are
	// read from the logs, which could not be.
	Unknown string `json:"unknown,omitempty"`
	// Problem is one of the settings: OBI on, but nothing that lets it run.
	Problem string               `json:"problem,omitempty"`
	Nodes   []*OBINodeSyncOutput `json:"nodes,omitempty"`
}

// OBINodeSyncOutput is one node's check.
type OBINodeSyncOutput struct {
	NodeID   string `json:"nodeId"`
	Hostname string `json:"hostname,omitempty"`
	// Expected says the settings have the node run OBI.
	Expected bool `json:"expected"`
	// LastSeen is the node's latest status row, nil when none was found.
	LastSeen *time.Time    `json:"lastSeen,omitempty"`
	Wanted   bool          `json:"wanted"`
	Running  bool          `json:"running"`
	Action   OBISyncAction `json:"action"`
	Problem  string        `json:"problem,omitempty"`
}

type DBCleanupOutput struct {
	Error string `json:"error,omitempty"`
	// OrphanedAppsDeleted are the apps removed because their project or env no
	// longer exists.
	OrphanedAppsDeleted []*OrphanedAppOutput `json:"orphanedAppsDeleted,omitempty"`
}

type OrphanedAppOutput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Error string `json:"error,omitempty"`
}

type ClusterCleanupOutput struct {
	Nodes []*ClusterNodeCleanupOutput `json:"nodes"`
}

type ClusterNodeCleanupOutput struct {
	NodeID   string `json:"nodeId"`
	NodeName string `json:"nodeName"`

	ImagesDeleted         int    `json:"imagesDeleted"`
	ImagesPruneError      string `json:"imagesPruneError,omitempty"`
	VolumesDeleted        int    `json:"volumesDeleted"`
	VolumesPruneError     string `json:"volumesPruneError,omitempty"`
	ContainersDeleted     int    `json:"containersDeleted"`
	ContainersPruneError  string `json:"containersPruneError,omitempty"`
	TempContainersDeleted int    `json:"tempContainersDeleted"`
	TempServicesDeleted   int    `json:"tempServicesDeleted"`
	NetworksDeleted       int    `json:"networksDeleted"`
	NetworksPruneError    string `json:"networksPruneError,omitempty"`
	BuildCachesDeleted    int    `json:"buildCachesDeleted"`
	BuildCachesPruneError string `json:"buildCachesPruneError,omitempty"`
	SpaceReclaimed        uint64 `json:"spaceReclaimed"`
}

type CacheCleanupOutput struct {
	Error                   string `json:"error,omitempty"`
	RepoCacheFilesDeleted   int    `json:"repoCacheFilesDeleted"`
	RepoCacheSpaceReclaimed uint64 `json:"repoCacheSpaceReclaimed"`
}

type FileCleanupOutput struct {
	Error string `json:"error,omitempty"`
}

func (t *Task) OutputAsSystemCleanup() (*TaskSystemCleanupOutput, error) {
	return parseTaskOutputAs(t, func() *TaskSystemCleanupOutput { return &TaskSystemCleanupOutput{} })
}
