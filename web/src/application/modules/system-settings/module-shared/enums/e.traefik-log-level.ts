export const ETraefikLogLevel = {
    Trace: "TRACE",
    Debug: "DEBUG",
    Info: "INFO",
    Warn: "WARN",
    Error: "ERROR",
    Fatal: "FATAL",
    Panic: "PANIC",
} as const;

export type ETraefikLogLevel = (typeof ETraefikLogLevel)[keyof typeof ETraefikLogLevel];
