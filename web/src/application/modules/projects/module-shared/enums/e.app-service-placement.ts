export const EAppServicePlacement = {
    Equal: "==",
    NotEqual: "!=",
} as const;

export type EAppServicePlacement = (typeof EAppServicePlacement)[keyof typeof EAppServicePlacement];
