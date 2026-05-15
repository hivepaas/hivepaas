import { z } from "zod";

import { EEmailKind } from "@application/shared/enums";

export const EMAIL_HTTP_METHODS = ["GET", "POST", "PUT"] as const;

const KeyValueSchema = z.object({
    key: z.string(),
    value: z.string(),
});

export const CreateOrEditEmailAccountFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        kind: z.enum([EEmailKind.SMTP, EEmailKind.HTTP]),
        smtpHost: z.string().trim(),
        smtpPort: z.coerce.number().int().min(1, "Port must be at least 1").max(65535, "Port must be at most 65535"),
        smtpUsername: z.string().trim(),
        smtpDisplayName: z.string().trim(),
        smtpPassword: z.string().trim(),
        httpEndpoint: z.string().trim(),
        httpMethod: z.enum(EMAIL_HTTP_METHODS),
        httpUsername: z.string().trim(),
        httpDisplayName: z.string().trim(),
        httpPassword: z.string().trim(),
        httpContentType: z.string().trim(),
        headers: z.array(KeyValueSchema),
        fieldMapping: z.array(KeyValueSchema),
        availableInProjects: z.boolean(),
        default: z.boolean(),
    })
    .superRefine((value, ctx) => {
        if (value.kind === EEmailKind.SMTP) {
            if (!value.smtpHost) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["smtpHost"],
                    message: "Host is required",
                });
            }
            if (!value.smtpUsername) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["smtpUsername"],
                    message: "Username is required",
                });
            }
            if (!value.smtpPassword) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["smtpPassword"],
                    message: "Password is required",
                });
            }
        }

        if (value.kind === EEmailKind.HTTP) {
            if (!value.httpEndpoint) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    path: ["httpEndpoint"],
                    message: "Endpoint is required",
                });
            }
        }
    });

export type CreateOrEditEmailAccountFormInput = z.input<typeof CreateOrEditEmailAccountFormSchema>;
export type CreateOrEditEmailAccountFormOutput = z.output<typeof CreateOrEditEmailAccountFormSchema>;
