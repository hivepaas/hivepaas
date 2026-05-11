import { z } from "zod";

export const CreateOrEditBasicAuthFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    username: z.string().trim().min(1, "Username is required"),
    password: z.string().min(1, "Password is required"),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditBasicAuthFormInput = z.input<typeof CreateOrEditBasicAuthFormSchema>;
export type CreateOrEditBasicAuthFormOutput = z.output<typeof CreateOrEditBasicAuthFormSchema>;
