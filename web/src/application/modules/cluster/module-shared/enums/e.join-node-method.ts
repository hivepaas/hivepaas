export const EJoinNodeMethod = {
    RunCommandManually: "run-command-manually",
    RunCommandViaSSH: "run-command-via-ssh",
} as const;

export type EJoinNodeMethod = (typeof EJoinNodeMethod)[keyof typeof EJoinNodeMethod];
