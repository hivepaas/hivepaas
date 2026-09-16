package hperrors

// Errors for configuration spec export
var (
	ErrSpecMountTargetDuplicated   = NewErr(ErrValueInvalid, "ERR_SPEC_MOUNT_TARGET_DUPLICATED")
	ErrSpecSettingTypeUnclassified = NewErr(ErrInternal, "ERR_SPEC_SETTING_TYPE_UNCLASSIFIED")
)
