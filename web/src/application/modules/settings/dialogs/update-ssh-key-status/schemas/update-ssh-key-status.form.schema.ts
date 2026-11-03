import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateSSHKeyStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    inheritable: z.boolean(),
    default: z.boolean(),
});

export type UpdateSSHKeyStatusFormInput = z.input<typeof UpdateSSHKeyStatusFormSchema>;
export type UpdateSSHKeyStatusFormOutput = z.output<typeof UpdateSSHKeyStatusFormSchema>;
