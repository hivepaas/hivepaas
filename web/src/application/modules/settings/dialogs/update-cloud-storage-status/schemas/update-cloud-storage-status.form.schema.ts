import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const UpdateCloudStorageStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type UpdateCloudStorageStatusFormInput = z.input<typeof UpdateCloudStorageStatusFormSchema>;
export type UpdateCloudStorageStatusFormOutput = z.output<typeof UpdateCloudStorageStatusFormSchema>;
