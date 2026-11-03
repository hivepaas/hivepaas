export const ESeccompMode = {
    Default: "default",
    Unconfined: "unconfined",
    Custom: "custom",
} as const;

export type ESeccompMode = (typeof ESeccompMode)[keyof typeof ESeccompMode];
