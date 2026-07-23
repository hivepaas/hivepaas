import { z } from "zod";

export const EnvVarFormItemSchema = z.object({
    key: z.string().min(1, "Key is required"),
    value: z.string().min(1, "Value is required"),
    isLiteral: z.boolean(),
    isSystem: z.boolean(),
    isReadOnly: z.boolean(),
});

export const EnvVarsFormBaseSchema = z.object({
    buildtime: z.array(EnvVarFormItemSchema),
    runtime: z.array(EnvVarFormItemSchema),
});

export type EnvVarFormItemSchemaInput = z.input<typeof EnvVarFormItemSchema>;
export type EnvVarFormItemSchemaOutput = z.output<typeof EnvVarFormItemSchema>;
export type EnvVarsFormBaseSchemaInput = z.input<typeof EnvVarsFormBaseSchema>;
export type EnvVarsFormBaseSchemaOutput = z.output<typeof EnvVarsFormBaseSchema>;
