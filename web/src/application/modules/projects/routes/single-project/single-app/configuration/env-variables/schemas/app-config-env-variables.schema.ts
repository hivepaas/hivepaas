import { z } from "zod";
import { EnvVarFormItemSchema, EnvVarsFormBaseSchema } from "~/projects/module-shared/schemas";

export const AppConfigEnvVarsFormSchema = EnvVarsFormBaseSchema.extend({
    shared: z.array(EnvVarFormItemSchema),
});

export type AppConfigEnvVarsFormSchemaInput = z.input<typeof AppConfigEnvVarsFormSchema>;
export type AppConfigEnvVarsFormSchemaOutput = z.output<typeof AppConfigEnvVarsFormSchema>;
