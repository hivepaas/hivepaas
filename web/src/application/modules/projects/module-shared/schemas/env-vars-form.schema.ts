import { z } from "zod";

export const EnvVarsFormBaseSchema = z.object({
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

export type EnvVarsFormBaseSchemaInput = z.input<typeof EnvVarsFormBaseSchema>;
export type EnvVarsFormBaseSchemaOutput = z.output<typeof EnvVarsFormBaseSchema>;
