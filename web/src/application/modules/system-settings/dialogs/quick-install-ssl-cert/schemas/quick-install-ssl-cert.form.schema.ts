import { z } from "zod";

import { ESslCertType, ESslKeyType } from "@application/shared/enums";

const NamedObjectSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .nullish();

export const QuickInstallSslCertFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        domain: z.string().trim().min(1, "Domain is required"),
        wildcardDomain: z.boolean(),
        certType: z.nativeEnum(ESslCertType),
        provider: NamedObjectSchema,
        acmeProvider: NamedObjectSchema,
        email: z.string().trim().email("Invalid email"),
        keyType: z.nativeEnum(ESslKeyType),
        autoRenew: z.boolean(),
        certificate: z.string().trim(),
        privateKey: z.string().trim(),
        expireAt: z.date().optional().nullable(),
        notifyFrom: z.date().optional().nullable(),
    })
    .superRefine((value, ctx) => {
        const isCustom = value.certType === ESslCertType.Custom;
        const requiresProvider = value.certType === ESslCertType.ZeroSSL || value.certType === ESslCertType.GoogleTrust;

        if (requiresProvider && !value.provider?.id) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["provider"],
                message: "SSL Provider is required",
            });
        }

        if (value.wildcardDomain && !isCustom && !value.acmeProvider?.id) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["acmeProvider"],
                message: "ACME DNS Provider is required",
            });
        }

        if (!isCustom) {
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
