import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateAppHealthCheckStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().nullable().optional(),
    default: z.boolean(),
});

export type UpdateAppHealthCheckStatusFormInput = z.input<typeof UpdateAppHealthCheckStatusFormSchema>;
export type UpdateAppHealthCheckStatusFormOutput = z.output<typeof UpdateAppHealthCheckStatusFormSchema>;
