import { z } from "zod";

const CreateProjectAppFormSchemaBase = z.object({
    name: z
        .string({
            required_error: "Name is required",
        })
        .trim()
        .min(1, "Name is required"),
    env: z.string().max(50, "Environment must be at most 50 characters"),
    note: z.string().refine(val => val === "" || (val.length >= 10 && val.length <= 10000), {
        message: "Note must be empty or between 10 and 10000 characters",
    }),
    tags: z.array(z.string()),
});

export function createCreateProjectAppFormSchema(
    envNames: string[] = [],
): z.ZodEffects<typeof CreateProjectAppFormSchemaBase> {
    return CreateProjectAppFormSchemaBase.superRefine((values, ctx) => {
        if (values.env !== "" && !envNames.includes(values.env)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["env"],
                message: "Environment must be one of the project environments",
            });
        }
    });
}

export const CreateProjectAppFormSchema = createCreateProjectAppFormSchema();

export type CreateProjectAppFormInput = z.input<ReturnType<typeof createCreateProjectAppFormSchema>>;
export type CreateProjectAppFormOutput = z.output<ReturnType<typeof createCreateProjectAppFormSchema>>;
