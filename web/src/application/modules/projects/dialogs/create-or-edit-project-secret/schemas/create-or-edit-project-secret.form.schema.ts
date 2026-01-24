import { z } from "zod";

export const CreateOrEditProjectSecretFormSchema = z.object({
    name: z
        .string({
            required_error: "Name is required",
        })
        .trim()
        .min(1, "Name is required"),
    value: z
        .string({
            required_error: "Value is required",
        })
        .trim()
        .min(1, "Value is required"),
});

export type CreateOrEditProjectSecretFormInput = z.input<typeof CreateOrEditProjectSecretFormSchema>;
export type CreateOrEditProjectSecretFormOutput = z.output<typeof CreateOrEditProjectSecretFormSchema>;
