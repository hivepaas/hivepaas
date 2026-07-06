import { z } from "zod";

export const CreateOrEditAccessTokenFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    kind: z.string().trim().min(1, "Type is required"),
    user: z.string().trim().min(1, "User is required"),
    token: z.string().min(1, "Token is required"),
    baseURL: z.string().trim(),
    expireAt: z.date().nullable(),
    availableInProjects: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditAccessTokenFormInput = z.input<typeof CreateOrEditAccessTokenFormSchema>;
export type CreateOrEditAccessTokenFormOutput = z.output<typeof CreateOrEditAccessTokenFormSchema>;
