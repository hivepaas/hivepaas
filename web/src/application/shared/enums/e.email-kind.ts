export const EEmailKind = {
    SMTP: "smtp",
    HTTP: "http",
} as const;

export type EEmailKind = (typeof EEmailKind)[keyof typeof EEmailKind];
