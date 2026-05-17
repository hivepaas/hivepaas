import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateAccessTokenStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateAccessTokenStatusFormInput = z.input<typeof UpdateAccessTokenStatusFormSchema>;
export type UpdateAccessTokenStatusFormOutput = z.output<typeof UpdateAccessTokenStatusFormSchema>;
