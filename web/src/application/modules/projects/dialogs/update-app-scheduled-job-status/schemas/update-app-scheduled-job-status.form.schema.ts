import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateAppScheduledJobStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().nullable().optional(),
    default: z.boolean(),
});

export type UpdateAppScheduledJobStatusFormInput = z.input<typeof UpdateAppScheduledJobStatusFormSchema>;
export type UpdateAppScheduledJobStatusFormOutput = z.output<typeof UpdateAppScheduledJobStatusFormSchema>;
