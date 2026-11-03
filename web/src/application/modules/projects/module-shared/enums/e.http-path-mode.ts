export const EHttpPathMode = {
    Exact: "exact",
    Prefix: "prefix",
    Regex: "regex",
} as const;

export type EHttpPathMode = (typeof EHttpPathMode)[keyof typeof EHttpPathMode];
