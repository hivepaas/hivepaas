import { z } from "zod";

import { ECloudStorageKind } from "@application/shared/enums";

export const CreateOrEditCloudStorageFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    kind: z.nativeEnum(ECloudStorageKind),
    accessKeyId: z.string().trim().min(1, "Access Key ID is required"),
    secretKey: z.string().min(1, "Secret Key is required"),
    region: z.string().trim().min(1, "Region is required"),
    bucket: z.string().trim().min(1, "Bucket is required"),
    endpoint: z.string().trim(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditCloudStorageFormInput = z.input<typeof CreateOrEditCloudStorageFormSchema>;
export type CreateOrEditCloudStorageFormOutput = z.output<typeof CreateOrEditCloudStorageFormSchema>;
