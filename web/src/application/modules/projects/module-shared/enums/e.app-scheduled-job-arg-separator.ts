export const EAppScheduledJobArgSeparator = {
    Equal: "=",
    Whitespace: " ",
} as const;

export type EAppScheduledJobArgSeparator =
    (typeof EAppScheduledJobArgSeparator)[keyof typeof EAppScheduledJobArgSeparator];
