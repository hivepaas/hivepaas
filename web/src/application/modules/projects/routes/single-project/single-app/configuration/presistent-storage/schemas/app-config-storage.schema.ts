import { z } from "zod";
import { EMountConsistency, EMountType } from "~/projects/module-shared/enums";

export const StorageMountFormSchema = z.object({
    type: z.nativeEnum(EMountType).optional(),
    source: z.string(),
    target: z.string(),
    consistency: z.nativeEnum(EMountConsistency).optional(),
});

export const AppConfigStorageFormSchema = z.object({
    mounts: z.array(StorageMountFormSchema),
});

export type AppConfigStorageFormSchemaInput = z.input<typeof AppConfigStorageFormSchema>;
export type AppConfigStorageFormSchemaOutput = z.output<typeof AppConfigStorageFormSchema>;

export const emptyAppConfigStorageFormDefaults: AppConfigStorageFormSchemaInput = {
    mounts: [],
};
