import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

/**
 * The interval is from 1 to 10 hours: an Amazon ECR token lives 12, and one
 * handed to Swarm must outlive the next run by an hour.
 */
export const REGISTRY_AUTH_RENEWAL_INTERVALS = ["1h", "2h", "3h", "4h", "5h", "6h", "7h", "8h", "9h", "10h"] as const;

const SettingsRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const NotificationSchema = z.object({
    successUseDefault: z.boolean(),
    success: SettingsRefSchema.nullish(),
    failureUseDefault: z.boolean(),
    failure: SettingsRefSchema.nullish(),
});

export const SystemRegistryAuthRenewalConfigurationFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    scheduleInterval: z.string().min(1, "Required"),
    scheduleFrom: z.date().nullable(),
    notification: NotificationSchema,
});

export type SystemRegistryAuthRenewalConfigurationFormInput = z.input<
    typeof SystemRegistryAuthRenewalConfigurationFormSchema
>;
export type SystemRegistryAuthRenewalConfigurationFormOutput = z.output<
    typeof SystemRegistryAuthRenewalConfigurationFormSchema
>;
