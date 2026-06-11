import type { EAppScheduledJobTaskPriority, EAppScheduledJobTaskStatus } from "~/projects/module-shared/enums";

export type AppScheduledJobTaskLogFrameType = "in" | "out" | "err" | "warn" | "debug";

export interface AppScheduledJobTaskLogFrame {
    type: AppScheduledJobTaskLogFrameType;
    data: string;
    ts: Date | null;
}

export interface AppScheduledJobTaskConfig {
    priority: EAppScheduledJobTaskPriority;
    maxRetry: number;
    retry: number;
    retryDelay: string;
    timeout: string;
    controlDisabled: boolean;
}

export interface AppScheduledJobTask {
    id: string;
    type: string;
    status: EAppScheduledJobTaskStatus;
    config: AppScheduledJobTaskConfig;
    lastError: string;
    updateVer: number;
    runAt: Date | null;
    retryAt: Date | null;
    startedAt: Date | null;
    endedAt: Date | null;
    createdAt: Date;
    updatedAt: Date;
}
