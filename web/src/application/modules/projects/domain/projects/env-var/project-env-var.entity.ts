export type ProjectBuildtimeEnvVar = {
    key: string;
    value: string;
    isLiteral: boolean;
};

export type ProjectRuntimeEnvVar = {
    key: string;
    value: string;
    isLiteral: boolean;
};

export type ProjectEnvVar = {
    buildtime: ProjectBuildtimeEnvVar[];
    runtime: ProjectRuntimeEnvVar[];
    updateVer: number;
};
