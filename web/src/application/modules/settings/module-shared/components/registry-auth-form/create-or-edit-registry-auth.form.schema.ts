import { z } from "zod";

export const CreateOrEditRegistryAuthFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    address: z.string().trim().min(1, "Server address is required"),
    username: z.string().trim().min(1, "Username is required"),
    // 8 KB, the API's: a Google Artifact Registry JSON key in base64 is some 3 KB.
    password: z
        .string()
        .min(1, "Password is required")
        .max(8 * 1024, "At most 8 KB"),
    readonly: z.boolean(),
    inheritable: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditRegistryAuthFormInput = z.input<typeof CreateOrEditRegistryAuthFormSchema>;
export type CreateOrEditRegistryAuthFormOutput = z.output<typeof CreateOrEditRegistryAuthFormSchema>;
