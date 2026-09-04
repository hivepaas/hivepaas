import { z } from "zod";

export const ChangeKekFormSchema = z
    .object({
        currentSecret: z.string(),
        newSecret: z
            .string()
            .min(10, "New secret must be at least 10 characters long")
            .regex(/[A-Z]/, "New secret must contain at least one uppercase letter")
            .regex(/[a-z]/, "New secret must contain at least one lowercase letter")
            .regex(/[0-9]/, "New secret must contain at least one digit")
            .regex(/[^A-Za-z0-9]/, "New secret must contain at least one symbol"),
        confirmNewSecret: z.string(),
        isSaved: z.boolean(),
    })
    .superRefine((arg, ctx) => {
        if (arg.newSecret !== arg.confirmNewSecret) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Secrets do not match",
                path: ["confirmNewSecret"],
            });
        }
    });

export type ChangeKekFormSchemaInput = z.input<typeof ChangeKekFormSchema>;
export type ChangeKekFormSchemaOutput = z.output<typeof ChangeKekFormSchema>;
