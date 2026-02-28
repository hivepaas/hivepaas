import { z } from "zod";

export const AppConfigGeneralFormSchema = z.object({
    name: z.string().min(1, "Name is required"),
    tags: z.array(z.string()),
    note: z.string(),
});

export type AppConfigGeneralFormSchemaInput = z.input<typeof AppConfigGeneralFormSchema>;
export type AppConfigGeneralFormSchemaOutput = z.output<typeof AppConfigGeneralFormSchema>;
