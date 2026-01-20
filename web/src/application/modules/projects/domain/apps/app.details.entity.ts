import type { ProjectAppBase } from "./app.base.entity";

export interface ProjectAppDetails extends ProjectAppBase {
    key: string;
    photo: string | null;
    userAccesses: {
        id: string;
        fullName: string;
        access: {
            read: boolean;
            write: boolean;
            delete: boolean;
        };
    }[];
    updateVer: number;
    stats: {
        runningTasks: number;
        desiredTasks: number;
        completedTasks: number;
    };
}
