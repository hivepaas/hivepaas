import { z } from "zod";

export const APP_CONFIG_FILE_MAX_VALUE_SIZE = 1024 * 1024;
export const APP_CONFIG_FILE_DEFAULT_FILE_PATH = "/etc/myapp/config";
export const APP_CONFIG_FILE_DEFAULT_FILE_MODE = "0444";

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
        mountIntoFilesystem: z.boolean(),
        filePath: z.string(),
        fileMode: z.string(),
        fileUid: z.string(),
        fileGid: z.string(),
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

        if (value.mountIntoFilesystem) {
            if (!value.filePath.trim()) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: "File path is required",
                    path: ["filePath"],
                });
            }

            if (!value.fileMode.trim()) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: "File mode is required",
                    path: ["fileMode"],
                });
            }
        }
    });

export type CreateOrEditAppConfigFileFormInput = z.input<typeof CreateOrEditAppConfigFileFormSchema>;
export type CreateOrEditAppConfigFileFormOutput = z.output<typeof CreateOrEditAppConfigFileFormSchema>;
