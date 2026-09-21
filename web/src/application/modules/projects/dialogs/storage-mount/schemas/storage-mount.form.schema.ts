import { z } from "zod";
import { EMountConsistency } from "~/projects/module-shared/enums";

export const StorageMountFormSchema = z.object({
    source: z.string().min(1, "Volume is required"),
    subpath: z.string().optional(),
    readOnly: z.boolean().optional(),
    noCopy: z.boolean().optional(),
    target: z.string().min(1, "Target is required"),
    consistency: z.nativeEnum(EMountConsistency).optional(),
    /** Empty for the app's own directory, which is the ordinary case. */
    sourceAppId: z.string().optional(),
    sourceAppWrite: z.boolean().optional(),
});

export type StorageMountFormInput = z.input<typeof StorageMountFormSchema>;
export type StorageMountFormOutput = z.output<typeof StorageMountFormSchema>;
