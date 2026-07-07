import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateProjectCommandTemplateStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    default: z.boolean(),
});

export type UpdateProjectCommandTemplateStatusFormInput = z.input<typeof UpdateProjectCommandTemplateStatusFormSchema>;
export type UpdateProjectCommandTemplateStatusFormOutput = z.output<
    typeof UpdateProjectCommandTemplateStatusFormSchema
>;
