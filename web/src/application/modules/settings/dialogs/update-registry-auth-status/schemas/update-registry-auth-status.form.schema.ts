import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateRegistryAuthStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateRegistryAuthStatusFormInput = z.input<typeof UpdateRegistryAuthStatusFormSchema>;
export type UpdateRegistryAuthStatusFormOutput = z.output<typeof UpdateRegistryAuthStatusFormSchema>;
