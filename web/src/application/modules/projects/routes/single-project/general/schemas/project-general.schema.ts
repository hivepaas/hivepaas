import { z } from "zod";

export const ProjectGeneralFormSchema = z.object({
    name: z.string().min(1, "Name is required"),
    tags: z.array(z.string()).min(1, "At least one tag is required"),
    note: z.string().trim().min(10, "Note must be at least 10 characters long").max(10000, "Note must be less than 1000 characters long"),
});

export type ProjectGeneralFormSchemaInput = z.input<typeof ProjectGeneralFormSchema>;
export type ProjectGeneralFormSchemaOutput = z.output<typeof ProjectGeneralFormSchema>;

