export const EAppDeploymentMethod = {
    Repo: "repo",
    Image: "image",
    Function: "function",
} as const;

export type EAppDeploymentMethod = (typeof EAppDeploymentMethod)[keyof typeof EAppDeploymentMethod];
