export const EProjectAppStatus = {
    Active: "active",
    Locked: "locked",
    Disabled: "disabled",
    Deleting: "deleting",
} as const;

export type EProjectAppStatus = (typeof EProjectAppStatus)[keyof typeof EProjectAppStatus];
