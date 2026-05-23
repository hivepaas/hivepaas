import { z } from "zod";

const MAX_PROJECT_ENVS = 10;

const ProjectEnvFormSchema = z.object({
    name: z
        .string()
        .trim()
        .min(1, "Environment name is required")
        .max(50, "Environment name must be 50 characters or less"),
    color: z.string(),
});

export const ProjectGeneralFormSchema = z.object({
    name: z.string().min(1, "Name is required"),
    ownerId: z.string().min(1, "Project owner is required"),
    envs: z
        .array(ProjectEnvFormSchema)
        .max(MAX_PROJECT_ENVS, `Maximum ${MAX_PROJECT_ENVS} environments`)
        .superRefine((envs, ctx) => {
            const names = new Set<string>();
            envs.forEach((env, index) => {
                if (names.has(env.name)) {
                    ctx.addIssue({
                        code: z.ZodIssueCode.custom,
                        message: "Environment names must be unique",
                        path: [index, "name"],
                    });
                }
                names.add(env.name);
            });
        }),
    tags: z.array(z.string()),
    note: z.string(),
});

export type ProjectGeneralFormSchemaInput = z.input<typeof ProjectGeneralFormSchema>;
export type ProjectGeneralFormSchemaOutput = z.output<typeof ProjectGeneralFormSchema>;
