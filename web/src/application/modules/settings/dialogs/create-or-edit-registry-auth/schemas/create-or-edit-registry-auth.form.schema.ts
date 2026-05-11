import { z } from "zod";

export const CreateOrEditRegistryAuthFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    address: z.string().trim().min(1, "Server address is required"),
    username: z.string().trim().min(1, "Username is required"),
    password: z.string().min(1, "Password is required"),
    readonly: z.boolean(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditRegistryAuthFormInput = z.input<typeof CreateOrEditRegistryAuthFormSchema>;
export type CreateOrEditRegistryAuthFormOutput = z.output<typeof CreateOrEditRegistryAuthFormSchema>;
