export const EAppDeploymentMethod = {
    Repo: "repo",
    Image: "image",
} as const;

export type EAppDeploymentMethod = (typeof EAppDeploymentMethod)[keyof typeof EAppDeploymentMethod];
