import type { ProjectAppBase, ProjectAppBaseRef } from "./app.base.entity";

export interface ProjectAppDetails extends ProjectAppBase {
    key: string;
    updateVer: number;
    /** What the app is - "function", "database" - empty when it declares no kind. */
    category: string;
    accessLinks: string[];
    parentApp: ProjectAppBaseRef | null;
    stats: {
        runningTasks: number;
        desiredTasks: number;
        completedTasks: number;
    } | null;
    childApps?: ProjectAppDetails[];
    logicalChildApps?: ProjectAppDetails[];
    subApps?: ProjectAppDetails[];
}
