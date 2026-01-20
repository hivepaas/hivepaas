import { z } from "zod";

export const CreateProjectAppFormSchema = z.object({
    name: z
        .string({
            required_error: "Name is required",
        })
        .trim()
        .min(1, "Name is required"),
    note: z.string(),
    tags: z.array(z.string()).min(1, "At least one tag is required"),
});

export type CreateProjectAppFormInput = z.input<typeof CreateProjectAppFormSchema>;
export type CreateProjectAppFormOutput = z.output<typeof CreateProjectAppFormSchema>;
