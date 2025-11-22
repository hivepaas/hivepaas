import { z } from "zod";

const SUPPORTED_FORMATS = ["image/jpg", "image/jpeg", "image/png", "image/webp", "image/gif"];

export const UploadPhotoFileSchema = z
    .instanceof(File)
    .nullable()
    .superRefine((file, ctx) => {
        if (file === null) return;

        if (!SUPPORTED_FORMATS.includes(file.type)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "File format should be one of the following: jpg, jpeg, png, webp, gif",
            });

            return z.NEVER;
        }
    });

export const UploadPhotoFormSchema = z.object({
    photo: UploadPhotoFileSchema,
});

export type UploadPhotoFormInput = z.input<typeof UploadPhotoFormSchema>;
export type UploadPhotoFormOutput = z.output<typeof UploadPhotoFormSchema>;
