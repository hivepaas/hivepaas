import { z } from "zod";

export const PROJECT_SECRET_MAX_VALUE_SIZE = 500 * 1024;

export const CreateOrEditProjectSecretFormSchema = z
    .object({
        name: z
            .string({
                required_error: "Name is required",
            })
            .trim()
            .min(1, "Name is required"),
        valueType: z.enum(["text", "binary"]),
        isEditMode: z.boolean(),
        textValue: z.string(),
        binaryFile: z.custom<File>().nullable(),
    })
    .superRefine((value, ctx) => {
        if (value.valueType === "text") {
            if (!value.textValue.trim() && !value.isEditMode) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: "Value is required",
                    path: ["textValue"],
                });
            }

            if (new TextEncoder().encode(value.textValue).byteLength > PROJECT_SECRET_MAX_VALUE_SIZE) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: "Value must be 500kb or less",
                    path: ["textValue"],
                });
            }
        }

        if (value.valueType === "binary" && !value.binaryFile && !value.isEditMode) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "File is required",
                path: ["binaryFile"],
            });
        }

        if (value.binaryFile && value.binaryFile.size > PROJECT_SECRET_MAX_VALUE_SIZE) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "File must be 500kb or less",
                path: ["binaryFile"],
            });
        }
    });

export type CreateOrEditProjectSecretFormInput = z.input<typeof CreateOrEditProjectSecretFormSchema>;
export type CreateOrEditProjectSecretFormOutput = z.output<typeof CreateOrEditProjectSecretFormSchema>;
