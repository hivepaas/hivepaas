import type { ProjectAppBase, ProjectAppBaseRef } from "./app.base.entity";

export interface ProjectAppDetails extends ProjectAppBase {
    key: string;
    updateVer: number;
    accessLinks: string[];
    parentApp: ProjectAppBaseRef | null;
    stats: {
        runningTasks: number;
        desiredTasks: number;
        completedTasks: number;
    } | null;
}
