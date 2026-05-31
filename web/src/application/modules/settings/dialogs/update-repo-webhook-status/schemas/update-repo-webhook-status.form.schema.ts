import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateRepoWebhookStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateRepoWebhookStatusFormInput = z.input<typeof UpdateRepoWebhookStatusFormSchema>;
export type UpdateRepoWebhookStatusFormOutput = z.output<typeof UpdateRepoWebhookStatusFormSchema>;
