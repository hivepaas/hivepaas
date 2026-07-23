export type ProjectBuildtimeEnvVar = {
    key: string;
    value: string;
    isLiteral: boolean;
    isSystem: boolean;
    isReadOnly: boolean;
};

export type ProjectRuntimeEnvVar = {
    key: string;
    value: string;
    isLiteral: boolean;
    isSystem: boolean;
    isReadOnly: boolean;
};

export type ProjectSharedEnvVar = ProjectRuntimeEnvVar;

export type ProjectEnvVar = {
    buildtime: ProjectBuildtimeEnvVar[];
    runtime: ProjectRuntimeEnvVar[];
    updateVer: number;
};

export type ProjectAppEnvVar = ProjectEnvVar & {
    shared: ProjectSharedEnvVar[];
};
