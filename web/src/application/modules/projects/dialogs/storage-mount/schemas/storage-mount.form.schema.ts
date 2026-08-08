import { z } from "zod";
import { EMountConsistency } from "~/projects/module-shared/enums";

export const StorageMountFormSchema = z.object({
    source: z.string().min(1, "Volume is required"),
    subpath: z.string().optional(),
    readOnly: z.boolean().optional(),
    noCopy: z.boolean().optional(),
    target: z.string().min(1, "Target is required"),
    consistency: z.nativeEnum(EMountConsistency).optional(),
});

export type StorageMountFormInput = z.input<typeof StorageMountFormSchema>;
export type StorageMountFormOutput = z.output<typeof StorageMountFormSchema>;
