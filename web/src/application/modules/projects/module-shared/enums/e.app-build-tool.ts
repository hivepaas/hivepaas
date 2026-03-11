export const EBuildTool = {
    Docker: "docker",
    Nixpacks: "nixpacks",
} as const;

export type EBuildTool = (typeof EBuildTool)[keyof typeof EBuildTool];
