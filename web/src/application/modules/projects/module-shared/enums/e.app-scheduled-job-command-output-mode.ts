export const EAppScheduledJobCommandOutputMode = {
    Ignore: "ignore",
    SaveToFile: "saveToFile",
    PipeToApp: "pipeToApp",
} as const;

export type EAppScheduledJobCommandOutputMode =
    (typeof EAppScheduledJobCommandOutputMode)[keyof typeof EAppScheduledJobCommandOutputMode];
