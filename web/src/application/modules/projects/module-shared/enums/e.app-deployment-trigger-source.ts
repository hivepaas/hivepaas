export const EAppDeploymentTriggerSource = {
    User: "user",
    RepoWebhook: "repo-webhook",
    API: "api",
} as const;

export type EAppDeploymentTriggerSource =
    (typeof EAppDeploymentTriggerSource)[keyof typeof EAppDeploymentTriggerSource];
