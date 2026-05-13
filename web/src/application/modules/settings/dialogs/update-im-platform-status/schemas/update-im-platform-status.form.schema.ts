import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateImPlatformStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateImPlatformStatusFormInput = z.input<typeof UpdateImPlatformStatusFormSchema>;
export type UpdateImPlatformStatusFormOutput = z.output<typeof UpdateImPlatformStatusFormSchema>;
