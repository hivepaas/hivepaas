import type { ProjectAppBase } from "./app.base.entity";

export interface ProjectAppDetails extends ProjectAppBase {
    key: string;
    localKey: string;
    updateVer: number;
    accessLinks: string[];
    stats: {
        runningTasks: number;
        desiredTasks: number;
        completedTasks: number;
    } | null;
}
