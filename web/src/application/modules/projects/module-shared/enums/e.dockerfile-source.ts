export const EDockerfileSource = {
    Manual: "manual",
    Auto: "auto",
} as const;

export type EDockerfileSource = (typeof EDockerfileSource)[keyof typeof EDockerfileSource];
