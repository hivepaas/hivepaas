package clustersecretserviceimpl

import (
	"time"
)

const (
	secretDefaultFileUID  = "0"
	secretDefaultFileGID  = "0"
	secretDefaultFileMode = 444
)

const (
	itemRemovalRetryMax   = 2
	itemRemovalRetryDelay = 2 * time.Second
	itemRemovalRetryIncr  = 1 * time.Second
)
