import { z } from "zod";

const LocalSchema = z.object({
    uploadMethod: z.literal("local"),
    files: z.array(z.instanceof(File)).min(1, "Select at least one file"),
    fileKind: z.string().trim().min(1, "File Kind is required"),
});

const CloudSchema = z.object({
    uploadMethod: z.literal("cloud"),
    storage: z
        .object({
            id: z.string(),
            name: z.string(),
        })
        .nullable(),
    bucket: z.string().trim().min(1, "Bucket is required"),
    fileKind: z.string().trim().min(1, "File Kind is required"),
    filePath: z.string().trim().min(1, "File Path is required"),
});

export const UploadAppDataFileFormSchema = z
    .discriminatedUnion("uploadMethod", [LocalSchema, CloudSchema])
    .superRefine((data, ctx) => {
        if (data.uploadMethod === "cloud" && data["storage"] === null) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Storage is required",
                path: ["storage"],
            });
        }
    });

export type UploadAppDataFileFormValues = z.infer<typeof UploadAppDataFileFormSchema>;
