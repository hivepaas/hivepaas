import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateSslCertStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateSslCertStatusFormInput = z.input<typeof UpdateSslCertStatusFormSchema>;
export type UpdateSslCertStatusFormOutput = z.output<typeof UpdateSslCertStatusFormSchema>;
