import { z } from "zod";

import { ESecuritySettings } from "@application/shared/enums";

const BaseSchema = z.object({
    username: z.string().trim().min(1, "Username is required"),
    fullName: z.string().trim().min(1, "Full Name is required"),
    email: z.string().trim().min(1, "Email Address is required").email(),
    password: z.string().trim().min(1, "Password is required"),
    photo: z
        .object({
            fileName: z.string(),
            dataBase64: z.string(),
        })
        .nullable(),
    agreeTermsAndConditions: z.boolean().superRefine((arg, ctx) => {
        if (!arg) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: "You must agree to the terms and conditions",
            });
        }
    }),

    securityOption: z.nativeEnum(ESecuritySettings),
    mfaTotpSecret: z.string().optional(),
    passcode: z.string().optional(),

    // extra
    isStrongPassword: z.boolean(),
});

export const SignUpFormSchema = BaseSchema.superRefine((arg, ctx) => {
    if (arg.securityOption === ESecuritySettings.Password2FA && !arg.passcode) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Passcode is required",
            path: ["passcode"],
        });
    }

    if (arg.password.length > 0 && !arg.isStrongPassword) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Password is not strong enough",
            path: ["isStrongPassword"],
        });
    }
});

export type SighUpFormSchemaInput = z.input<typeof SignUpFormSchema>;
export type SighUpFormSchemaOutput = z.output<typeof SignUpFormSchema>;
