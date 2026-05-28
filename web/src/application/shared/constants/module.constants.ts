export const MODULE_IDS = {
    Settings: "mod::settings",
    Cluster: "mod::cluster",
    User: "mod::user",
    Project: "mod::project",
    System: "mod::system",
} as const;

export type ResourceModuleId = (typeof MODULE_IDS)[keyof typeof MODULE_IDS];

export const MODULES = [
    {
        name: "User Management",
        id: MODULE_IDS.User,
    },
    {
        name: "Project Management",
        id: MODULE_IDS.Project,
    },
    {
        name: "Cluster Management",
        id: MODULE_IDS.Cluster,
    },
    {
        name: "Settings",
        id: MODULE_IDS.Settings,
    },
    {
        name: "System",
        id: MODULE_IDS.System,
    },
] as const;
