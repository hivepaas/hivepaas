export const ESchedJobSeqMode = {
    Sequential: "sequential",
} as const;

export type ESchedJobSeqMode = (typeof ESchedJobSeqMode)[keyof typeof ESchedJobSeqMode];

export const ESchedJobSeqOnFailure = {
    Stop: "stop",
    Continue: "continue",
} as const;

export type ESchedJobSeqOnFailure = (typeof ESchedJobSeqOnFailure)[keyof typeof ESchedJobSeqOnFailure];

export const ESchedJobSeqStepStatus = {
    Pending: "pending",
    Running: "running",
    Done: "done",
    Failed: "failed",
    Skipped: "skipped",
} as const;

export type ESchedJobSeqStepStatus = (typeof ESchedJobSeqStepStatus)[keyof typeof ESchedJobSeqStepStatus];
