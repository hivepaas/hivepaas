import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateSslProviderStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateSslProviderStatusFormInput = z.input<typeof UpdateSslProviderStatusFormSchema>;
export type UpdateSslProviderStatusFormOutput = z.output<typeof UpdateSslProviderStatusFormSchema>;
