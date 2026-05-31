export const EAppHealthCheckType = {
    REST: "rest",
    GRPC: "grpc",
} as const;

export type EAppHealthCheckType = (typeof EAppHealthCheckType)[keyof typeof EAppHealthCheckType];
