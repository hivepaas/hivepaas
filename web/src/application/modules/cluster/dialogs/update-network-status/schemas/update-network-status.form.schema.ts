import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateNetworkStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateNetworkStatusFormInput = z.input<typeof UpdateNetworkStatusFormSchema>;
export type UpdateNetworkStatusFormOutput = z.output<typeof UpdateNetworkStatusFormSchema>;
