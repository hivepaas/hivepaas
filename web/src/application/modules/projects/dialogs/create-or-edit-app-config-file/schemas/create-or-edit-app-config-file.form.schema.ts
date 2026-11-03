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
        /** The type of the content being edited; null for a new config file. */
        initialValueType: z.enum(["text", "binary"]).nullable(),
        textValue: z.string(),
        binaryFile: z.custom<File>().nullable(),
        inheritable: z.boolean(),
    })
    .superRefine((value, ctx) => {
        // A text config file is edited in place, so its value is always there: an
        // emptied one would be read as "keep the content" by the backend. Only a
        // binary file left as it was keeps its content without a new upload.
        const keepsBinary = value.isEditMode && value.initialValueType === "binary";

        if (value.valueType === "text") {
            if (!value.textValue.trim()) {
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

        if (value.valueType === "binary" && !value.binaryFile && !keepsBinary) {
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
