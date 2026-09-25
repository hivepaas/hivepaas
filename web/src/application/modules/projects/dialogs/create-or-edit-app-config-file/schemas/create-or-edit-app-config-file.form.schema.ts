import { z } from "zod";

export const APP_CONFIG_FILE_MAX_VALUE_SIZE = 1024 * 1024;

export const CreateOrEditAppConfigFileFormSchema = z
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

            if (new TextEncoder().encode(value.textValue).byteLength > APP_CONFIG_FILE_MAX_VALUE_SIZE) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: "Value must be 1mb or less",
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

        if (value.binaryFile && value.binaryFile.size > APP_CONFIG_FILE_MAX_VALUE_SIZE) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "File must be 1mb or less",
                path: ["binaryFile"],
            });
        }
    });

export type CreateOrEditAppConfigFileFormInput = z.input<typeof CreateOrEditAppConfigFileFormSchema>;
export type CreateOrEditAppConfigFileFormOutput = z.output<typeof CreateOrEditAppConfigFileFormSchema>;
