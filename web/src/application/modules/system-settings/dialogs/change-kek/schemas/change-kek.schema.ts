import { z } from "zod";

export const ChangeKekFormSchema = z
    .object({
        currentSecret: z.string(),
        newSecret: z.string().trim().nonempty("New secret is required"),
        confirmNewSecret: z.string().trim().nonempty("Confirm new secret is required"),
        isSaved: z.boolean(),
        isStrongSecret: z.boolean(),
    })
    .superRefine((arg, ctx) => {
        if (arg.newSecret !== arg.confirmNewSecret) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Secrets do not match",
                path: ["confirmNewSecret"],
            });
        }

        if (arg.newSecret.length > 0 && !arg.isStrongSecret) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Secret is not strong enough",
                path: ["isStrongSecret"],
            });
        }
    });

export type ChangeKekFormSchemaInput = z.input<typeof ChangeKekFormSchema>;
export type ChangeKekFormSchemaOutput = z.output<typeof ChangeKekFormSchema>;
