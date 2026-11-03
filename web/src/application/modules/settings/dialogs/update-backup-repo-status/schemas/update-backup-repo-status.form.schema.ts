import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateBackupRepoStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    inheritable: z.boolean(),
    default: z.boolean(),
});

export type UpdateBackupRepoStatusFormInput = z.input<typeof UpdateBackupRepoStatusFormSchema>;
export type UpdateBackupRepoStatusFormOutput = z.output<typeof UpdateBackupRepoStatusFormSchema>;
