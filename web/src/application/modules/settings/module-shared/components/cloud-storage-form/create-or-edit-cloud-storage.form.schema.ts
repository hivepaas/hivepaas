import { z } from "zod";

import { ECloudStorageKind } from "@application/shared/enums";

export const CreateOrEditCloudStorageFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    kind: z.nativeEnum(ECloudStorageKind),
    keyAuth: z
        .object({ id: z.string(), name: z.string() })
        .nullable()
        .refine((value): boolean => value !== null, "Key auth is required"),
    region: z.string().trim().min(1, "Region is required"),
    bucket: z.string().trim().min(1, "Bucket is required"),
    endpoint: z.string().trim(),
    inheritable: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditCloudStorageFormInput = z.input<typeof CreateOrEditCloudStorageFormSchema>;
export type CreateOrEditCloudStorageFormOutput = z.output<typeof CreateOrEditCloudStorageFormSchema>;
