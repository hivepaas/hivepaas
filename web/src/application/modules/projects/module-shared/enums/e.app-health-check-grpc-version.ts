export const EAppHealthCheckGrpcVersion = {
    V1: "v1",
} as const;

export type EAppHealthCheckGrpcVersion = (typeof EAppHealthCheckGrpcVersion)[keyof typeof EAppHealthCheckGrpcVersion];
