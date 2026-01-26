import { z } from "zod";

export const ProjectEnvVarsFormSchema = z.object({
    buildtime: z.array(
        z.object({
            key: z.string().min(1, "Key is required"),
            value: z.string().min(1, "Value is required"),
            isLiteral: z.boolean(),
        }),
    ),
    runtime: z.array(
        z.object({
            key: z.string().min(1, "Key is required"),
            value: z.string().min(1, "Value is required"),
            isLiteral: z.boolean(),
        }),
    ),
});

export type ProjectEnvVarsFormSchemaInput = z.input<typeof ProjectEnvVarsFormSchema>;
export type ProjectEnvVarsFormSchemaOutput = z.output<typeof ProjectEnvVarsFormSchema>;
