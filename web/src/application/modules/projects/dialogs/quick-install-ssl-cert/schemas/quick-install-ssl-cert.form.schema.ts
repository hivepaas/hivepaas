import { z } from "zod";

import { ESslCertType, ESslKeyType } from "@application/shared/enums";

export const QuickInstallSslCertFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        domain: z.string().trim().min(1, "Domain is required"),
        certType: z.nativeEnum(ESslCertType),
        email: z.string().trim().email("Invalid email"),
        keyType: z.nativeEnum(ESslKeyType),
        autoRenew: z.boolean(),
        certificate: z.string().trim(),
        privateKey: z.string().trim(),
        expireAt: z.date().optional().nullable(),
        notifyFrom: z.date().optional().nullable(),
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

export type QuickInstallSslCertFormInput = z.input<typeof QuickInstallSslCertFormSchema>;
export type QuickInstallSslCertFormOutput = z.output<typeof QuickInstallSslCertFormSchema>;
