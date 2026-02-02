import { z } from "zod";

export const ProjectGeneralFormSchema = z.object({
    name: z.string().min(1, "Name is required"),
    tags: z.array(z.string()),
    note: z.string(),
});

export type ProjectGeneralFormSchemaInput = z.input<typeof ProjectGeneralFormSchema>;
export type ProjectGeneralFormSchemaOutput = z.output<typeof ProjectGeneralFormSchema>;
