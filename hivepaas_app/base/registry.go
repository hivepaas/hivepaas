package base

// RegistryType names the registry software, independently of who runs it.
type RegistryType string

const (
	RegistryTypeZot RegistryType = "zot"
)

// RegistryStorageType is where the images are kept. It is decided when the
// registry is provisioned and refused afterwards: nothing copies images from one
// to the other.
type RegistryStorageType string

const (
	RegistryStorageTypeVolume RegistryStorageType = "volume"
	RegistryStorageTypeS3     RegistryStorageType = "s3"
)

// RegistryCleanupMode says how the set of images to keep is decided.
//
// "policy" is the only value today: two numbers, turned into zot's own retention
// rules. A second value would be one where HivePaaS computes the set from what is
// running and tells the registry exactly which tags to keep.
type RegistryCleanupMode string

const (
	RegistryCleanupModePolicy RegistryCleanupMode = "policy"
)
