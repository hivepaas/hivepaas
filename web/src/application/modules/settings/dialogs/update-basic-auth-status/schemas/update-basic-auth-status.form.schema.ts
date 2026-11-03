import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateBasicAuthStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    inheritable: z.boolean(),
    default: z.boolean(),
});

export type UpdateBasicAuthStatusFormInput = z.input<typeof UpdateBasicAuthStatusFormSchema>;
export type UpdateBasicAuthStatusFormOutput = z.output<typeof UpdateBasicAuthStatusFormSchema>;
