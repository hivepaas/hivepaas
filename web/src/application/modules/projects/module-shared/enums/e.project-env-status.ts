export const EProjectEnvStatus = {
    Active: "active",
    Disabled: "disabled",
    Deleting: "deleting",
    Missing: "missing",
} as const;

export type EProjectEnvStatus = (typeof EProjectEnvStatus)[keyof typeof EProjectEnvStatus];
