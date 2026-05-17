import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateOAuthStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    default: z.boolean(),
});

export type UpdateOAuthStatusFormInput = z.input<typeof UpdateOAuthStatusFormSchema>;
export type UpdateOAuthStatusFormOutput = z.output<typeof UpdateOAuthStatusFormSchema>;
