import { z } from "zod";

export const CreateProjectFormSchema = z
    .object({
        name: z
            .string({
                required_error: "Name is required",
            })
            .trim()
            .min(1, "Name is required"),
        note: z.string().trim().min(10, "Note must be at least 10 characters long").max(10000, "Note must be less than 1000 characters long"),
        tags: z.array(z.string()).min(1, "At least one tag is required"),
    });

export type CreateProjectFormInput = z.input<typeof CreateProjectFormSchema>;
export type CreateProjectFormOutput = z.output<typeof CreateProjectFormSchema>;

