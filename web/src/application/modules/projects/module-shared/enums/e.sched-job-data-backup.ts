/** What a data backup reads: a command's output, or a volume the app mounts. */
export const ESchedJobDataBackupSource = {
    Command: "command",
    Volume: "volume",
} as const;

export type ESchedJobDataBackupSource = (typeof ESchedJobDataBackupSource)[keyof typeof ESchedJobDataBackupSource];
