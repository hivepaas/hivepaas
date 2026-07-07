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
            success: NamedObjectSchema,
            failureUseDefault: z.boolean(),
            failure: NamedObjectSchema,
        }),
    })
    .superRefine((value, ctx) => {
        const requiresAcmeProvider =
            value.domain.trim().startsWith("*.") &&
            value.certType !== ESslCertType.Custom &&
            value.certType !== ESslCertType.SelfSigned;

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
