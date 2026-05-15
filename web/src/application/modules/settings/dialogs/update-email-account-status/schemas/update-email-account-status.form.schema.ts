import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateEmailAccountStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateEmailAccountStatusFormInput = z.input<typeof UpdateEmailAccountStatusFormSchema>;
export type UpdateEmailAccountStatusFormOutput = z.output<typeof UpdateEmailAccountStatusFormSchema>;
