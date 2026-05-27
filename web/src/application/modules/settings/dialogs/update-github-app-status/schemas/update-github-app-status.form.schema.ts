import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateGithubAppStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateGithubAppStatusFormInput = z.input<typeof UpdateGithubAppStatusFormSchema>;
export type UpdateGithubAppStatusFormOutput = z.output<typeof UpdateGithubAppStatusFormSchema>;
