export const EAppDeploymentLogType = {
    In: "in",
    Out: "out",
    Err: "err",
    Warn: "warn",
    Debug: "debug",
} as const;

export type EAppDeploymentLogType = (typeof EAppDeploymentLogType)[keyof typeof EAppDeploymentLogType];

export interface AppDeploymentLogFrame {
    type: EAppDeploymentLogType;
    data: string;
    ts: Date | null;
}
