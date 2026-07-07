export const ECommandTemplateKind = {
    Backup: "backup",
    DataOps: "data-ops",
    Database: "database",
    Deployment: "deployment",
    Diagnostics: "diagnostics",
    Maintenance: "maintenance",
    Testing: "testing",
} as const;

export type ECommandTemplateKind = (typeof ECommandTemplateKind)[keyof typeof ECommandTemplateKind];
