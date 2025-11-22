import { z } from "zod";

export const AccountPasswordFormSchema = z
    .object({
        currentPassword: z.string().trim().nonempty("Current password is required"),
        newPassword: z.string().trim().nonempty("New password is required"),
        confirmNewPassword: z.string().trim().nonempty("Confirm new password is required"),

        // extra
        isStrongPassword: z.boolean(),
    })
    .superRefine((arg, ctx) => {
        if (arg.newPassword !== arg.confirmNewPassword) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Passwords do not match",
                path: ["confirmNewPassword"],
            });
        }

        if (arg.newPassword.length > 0 && !arg.isStrongPassword) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "Password is not strong enough",
                path: ["isStrongPassword"],
            });
        }
    });

export type AccountPasswordFormSchemaInput = z.input<typeof AccountPasswordFormSchema>;
export type AccountPasswordFormSchemaOutput = z.output<typeof AccountPasswordFormSchema>;
