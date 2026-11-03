import { z } from "zod";

import { ESslCertType, ESslKeyType } from "@application/shared/enums";

const NamedObjectSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .optional();

export const CreateOrEditSslCertFormSchema = z
    .object({
        domain: z.string().trim().min(1, "Domain is required"),
        certType: z.nativeEnum(ESslCertType),
        provider: NamedObjectSchema.nullish(),
        acmeProvider: NamedObjectSchema.nullish(),
        email: z.string().trim(),
        keyType: z.nativeEnum(ESslKeyType),
        autoRenew: z.boolean(),
        certificate: z.string().trim(),
        privateKey: z.string().trim(),
        caCertificate: z.string().trim().optional(),
        expireAt: z.date().optional().nullable(),

        notifyFrom: z.date().optional().nullable(),
        inheritable: z.boolean(),
        default: z.boolean(),
        notification: z.object({
            successUseDefault: z.boolean(),
            success: NamedObjectSchema.nullable(),
            failureUseDefault: z.boolean(),
            failure: NamedObjectSchema.nullable(),
        }),
    })
    .superRefine((value, ctx) => {
        // An email registers an ACME account; a certificate of one's own or one
        // signed here has none to register.
        const requiresEmail =
            value.certType === ESslCertType.LetsEncrypt ||
            value.certType === ESslCertType.ZeroSSL ||
            value.certType === ESslCertType.GoogleTrust;
        if ((requiresEmail || value.email) && !z.string().email().safeParse(value.email).success) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["email"],
                message: "Invalid email",
            });
        }

        const requiresProvider = value.certType === ESslCertType.ZeroSSL || value.certType === ESslCertType.GoogleTrust;
        const requiresAcmeProvider =
            value.domain.includes("*") &&
            value.certType !== ESslCertType.Custom &&
            value.certType !== ESslCertType.SelfSigned;

        if (requiresProvider && !value.provider?.id) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["provider"],
                message: "SSL Provider is required",
            });
        }

        if (requiresAcmeProvider && !value.acmeProvider?.id) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["acmeProvider"],
                message: "ACME DNS Provider is required",
            });
        }

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
    });

export type CreateOrEditSslCertFormInput = z.input<typeof CreateOrEditSslCertFormSchema>;
export type CreateOrEditSslCertFormOutput = z.output<typeof CreateOrEditSslCertFormSchema>;
