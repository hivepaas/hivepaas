import { z } from "zod";

export const CreateOrEditKeyAuthFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    keyId: z.string().trim().min(1, "Key ID is required"),
    secretKey: z.string().min(1, "Secret key is required"),
    inheritable: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditKeyAuthFormInput = z.input<typeof CreateOrEditKeyAuthFormSchema>;
export type CreateOrEditKeyAuthFormOutput = z.output<typeof CreateOrEditKeyAuthFormSchema>;
