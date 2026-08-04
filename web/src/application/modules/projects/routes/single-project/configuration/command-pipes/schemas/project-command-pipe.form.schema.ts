import { z } from "zod";

export const ProjectCommandPipeFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        sourceCommandId: z.string().trim(),
        targetCommandId: z.string().trim(),
        default: z.boolean(),
    })
    .superRefine((value, ctx) => {
        if (value.sourceCommandId || value.targetCommandId) {
            return;
        }

        const message = "Select at least one command (Source or Target)";

        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: ["sourceCommandId"],
            message,
        });
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: ["targetCommandId"],
            message,
        });
    });

export type ProjectCommandPipeFormInput = z.input<typeof ProjectCommandPipeFormSchema>;
export type ProjectCommandPipeFormOutput = z.output<typeof ProjectCommandPipeFormSchema>;
