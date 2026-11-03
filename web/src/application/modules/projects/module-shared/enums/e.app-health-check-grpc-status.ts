export const EAppHealthCheckGrpcStatus = {
    Unknown: 0,
    Serving: 1,
    NotServing: 2,
    ServiceUnknown: 3,
} as const;

export type EAppHealthCheckGrpcStatus = (typeof EAppHealthCheckGrpcStatus)[keyof typeof EAppHealthCheckGrpcStatus];
