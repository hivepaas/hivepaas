import { z } from "zod";

export const CreateProjectFormSchema = z
    .object({
        name: z
            .string({
                required_error: "Name is required",
            })
            .trim()
            .min(1, "Name is required"),
        note: z.string().trim(),
        tags: z.array(z.string()),
    });

export type CreateProjectFormInput = z.input<typeof CreateProjectFormSchema>;
export type CreateProjectFormOutput = z.output<typeof CreateProjectFormSchema>;

