export const EAppHealthCheckReturnBodyMode = {
    Skipped: "skipped",
    Text: "text",
    JSON: "json",
} as const;

export type EAppHealthCheckReturnBodyMode =
    (typeof EAppHealthCheckReturnBodyMode)[keyof typeof EAppHealthCheckReturnBodyMode];
