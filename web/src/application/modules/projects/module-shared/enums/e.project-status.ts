export const EProjectStatus = {
    Active: "active",
    Locked: "locked",
    Disabled: "disabled",
    Deleting: "deleting",
} as const;

export type EProjectStatus = (typeof EProjectStatus)[keyof typeof EProjectStatus];
