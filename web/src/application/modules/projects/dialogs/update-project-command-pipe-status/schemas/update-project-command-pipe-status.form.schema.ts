import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateProjectCommandPipeStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    default: z.boolean(),
});

export type UpdateProjectCommandPipeStatusFormInput = z.input<typeof UpdateProjectCommandPipeStatusFormSchema>;
export type UpdateProjectCommandPipeStatusFormOutput = z.output<typeof UpdateProjectCommandPipeStatusFormSchema>;
