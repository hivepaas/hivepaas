import { z } from "zod";

import { ESslCertType, ESslKeyType } from "@application/shared/enums";

const NotificationTargetSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .optional();

export const CreateOrEditSslCertFormSchema = z
    .object({
        domain: z.string().trim().min(1, "Domain is required"),
        certType: z.enum([ESslCertType.LetsEncrypt, ESslCertType.Custom]),
        email: z.string().trim().email("Invalid email"),
        keyType: z.nativeEnum(ESslKeyType),
        autoRenew: z.boolean(),
        certificate: z.string().trim(),
        privateKey: z.string().trim(),
        expireAt: z.date().optional().nullable(),
        notifyFrom: z.date().optional().nullable(),
        availableInProjects: z.boolean(),
        default: z.boolean(),
        notification: z.object({
            successUseDefault: z.boolean(),
            success: NotificationTargetSchema,
            failureUseDefault: z.boolean(),
            failure: NotificationTargetSchema,
        }),
    })
    .superRefine((value, ctx) => {
        if (value.certType !== ESslCertType.Custom) {
            return;
        }

        if (!value.certificate) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["certificate"],
                message: "Certificate is required",
            });
        }
        if (!value.privateKey) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["privateKey"],
                message: "Private key is required",
            });
        }
        if (!value.expireAt) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["expireAt"],
                message: "Expire At is required",
            });
        }
    });

export type CreateOrEditSslCertFormInput = z.input<typeof CreateOrEditSslCertFormSchema>;
export type CreateOrEditSslCertFormOutput = z.output<typeof CreateOrEditSslCertFormSchema>;
