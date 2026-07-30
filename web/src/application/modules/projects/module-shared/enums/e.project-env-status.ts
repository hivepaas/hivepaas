export const EProjectEnvStatus = {
    Active: "active",
    Disabled: "disabled",
    Deleting: "deleting",
} as const;

export type EProjectEnvStatus = (typeof EProjectEnvStatus)[keyof typeof EProjectEnvStatus];
