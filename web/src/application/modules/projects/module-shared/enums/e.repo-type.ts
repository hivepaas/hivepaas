export const ERepoType = {
    Git: "git",
} as const;

export type ERepoType = (typeof ERepoType)[keyof typeof ERepoType];
