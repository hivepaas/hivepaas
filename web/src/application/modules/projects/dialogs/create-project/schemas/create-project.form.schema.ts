import { z } from "zod";

export const DEFAULT_PROJECT_ENVS = [
    { name: "Development", color: "#a855f7" },
    { name: "Staging", color: "#eab308" },
    { name: "Production", color: "#84cc16" },
] as const;

export const MAX_PROJECT_ENVS = 10;

const ProjectEnvFormSchema = z.object({
    name: z
        .string({
            required_error: "Environment name is required",
        })
        .trim()
        .min(1, "Environment name is required")
        .max(50, "Environment name must be 50 characters or less"),
    color: z.string(),
});

export const CreateProjectFormSchema = z.object({
    name: z
        .string({
            required_error: "Name is required",
        })
        .trim()
        .min(1, "Name is required"),
    note: z.string().refine(val => val === "" || (val.length >= 10 && val.length <= 10000), {
        message: "Note must be empty or between 10 and 10000 characters",
    }),
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
});

export type CreateProjectFormInput = z.input<typeof CreateProjectFormSchema>;
export type CreateProjectFormOutput = z.output<typeof CreateProjectFormSchema>;
