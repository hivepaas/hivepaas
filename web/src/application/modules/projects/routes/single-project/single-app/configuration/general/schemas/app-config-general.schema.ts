import { z } from "zod";

const AppPhotoUploadSchema = z
    .union([
        z.object({
            fileName: z.string(),
            dataBase64: z.string(),
        }),
        z.object({
            delete: z.literal(true),
        }),
    ])
    .nullable();

export const AppConfigGeneralFormSchema = z.object({
    photo: z.string().nullable(),
    photoUpload: AppPhotoUploadSchema,
    name: z.string().min(1, "Name is required"),
    tags: z.array(z.string()),
    note: z.string(),
});

export type AppConfigGeneralFormSchemaInput = z.input<typeof AppConfigGeneralFormSchema>;
export type AppConfigGeneralFormSchemaOutput = z.output<typeof AppConfigGeneralFormSchema>;
