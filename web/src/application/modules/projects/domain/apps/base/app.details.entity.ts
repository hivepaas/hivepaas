import type { ProjectAppBase } from "./app.base.entity";

export interface ProjectAppDetails extends ProjectAppBase {
    key: string;
    localKey: string;
    updateVer: number;
    stats: {
        runningTasks: number;
        desiredTasks: number;
        completedTasks: number;
    } | null;
}
