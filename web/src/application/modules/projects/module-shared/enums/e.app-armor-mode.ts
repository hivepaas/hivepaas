export const EAppArmorMode = {
    Default: "default",
    Disabled: "disabled",
} as const;

export type EAppArmorMode = (typeof EAppArmorMode)[keyof typeof EAppArmorMode];
