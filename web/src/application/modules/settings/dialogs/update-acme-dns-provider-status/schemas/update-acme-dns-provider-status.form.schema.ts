import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateAcmeDnsProviderStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateAcmeDnsProviderStatusFormInput = z.input<typeof UpdateAcmeDnsProviderStatusFormSchema>;
export type UpdateAcmeDnsProviderStatusFormOutput = z.output<typeof UpdateAcmeDnsProviderStatusFormSchema>;
