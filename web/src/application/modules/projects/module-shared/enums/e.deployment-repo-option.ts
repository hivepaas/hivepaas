export const EDeploymentRepoOption = {
    GitSubmodulesEnabled: "gitSubmodulesEnabled",
    GitLfsEnabled: "gitLfsEnabled",
} as const;

export type EDeploymentRepoOption = (typeof EDeploymentRepoOption)[keyof typeof EDeploymentRepoOption];
