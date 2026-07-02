package entity

type TaskSystemCleanupOutput struct {
	DBCleanup      *DBCleanupOutput      `json:"dbCleanup"`
	ClusterCleanup *ClusterCleanupOutput `json:"clusterCleanup"`
	BackupCleanup  *BackupCleanupOutput  `json:"backupCleanup"`
	CacheCleanup   *CacheCleanupOutput   `json:"cacheCleanup"`
	FileCleanup    *FileCleanupOutput    `json:"fileCleanup"`
}

type DBCleanupOutput struct {
	Error string `json:"error,omitempty"`
}

type ClusterCleanupOutput struct {
	ImagesDeleted         int    `json:"imagesDeleted"`
	ImagesPruneError      string `json:"imagesPruneError,omitempty"`
	VolumesDeleted        int    `json:"volumesDeleted"`
	VolumesPruneError     string `json:"volumesPruneError,omitempty"`
	ContainersDeleted     int    `json:"containersDeleted"`
	ContainersPruneError  string `json:"containersPruneError,omitempty"`
	NetworksDeleted       int    `json:"networksDeleted"`
	NetworksPruneError    string `json:"networksPruneError,omitempty"`
	BuildCachesDeleted    int    `json:"buildCachesDeleted"`
	BuildCachesPruneError string `json:"buildCachesPruneError,omitempty"`
	SpaceReclaimed        uint64 `json:"spaceReclaimed"`
}

type BackupCleanupOutput struct {
	Error               string `json:"error,omitempty"`
	LocalBackupsDeleted int    `json:"localBackupsDeleted"`
	CloudBackupsDeleted int    `json:"cloudBackupsDeleted"`
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
