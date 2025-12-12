export const ENodeRole = {
    Manager: "manager",
    Worker: "worker",
} as const;

export type ENodeRole = (typeof ENodeRole)[keyof typeof ENodeRole];
