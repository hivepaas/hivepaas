import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateKeyAuthStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    inheritable: z.boolean(),
    default: z.boolean(),
});

export type UpdateKeyAuthStatusFormInput = z.input<typeof UpdateKeyAuthStatusFormSchema>;
export type UpdateKeyAuthStatusFormOutput = z.output<typeof UpdateKeyAuthStatusFormSchema>;
