export interface SettingInUseDialogProps {
    /** What the operator tried to delete, for the sentence at the top. */
    settingName?: string;
    /** The URL of the failed delete. The usage list hangs off it. */
    requestUrl?: string;
}

export interface SettingInUseDialogState {
    mode: "open" | "closed";
}

export interface SettingInUseDialogOptions {
    props: SettingInUseDialogProps;
}
