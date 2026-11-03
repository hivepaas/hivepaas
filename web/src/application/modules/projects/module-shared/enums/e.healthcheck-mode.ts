export const EHealthcheckMode = {
    Inherit: "",
    None: "NONE",
    Cmd: "CMD",
    CmdShell: "CMD-SHELL",
} as const;

export type EHealthcheckMode = (typeof EHealthcheckMode)[keyof typeof EHealthcheckMode];
