import { z } from "zod";

import { ESslKeyType, ESslProviderKind } from "@application/shared/enums";

export const SSL_PROVIDER_UNSPECIFIED_KEY_TYPE = "__unspecified";

export const CreateOrEditSslProviderFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        kind: z.nativeEnum(ESslProviderKind),
        email: z.string().trim().email("Invalid email"),
        defaultKeyType: z.nativeEnum(ESslKeyType).or(z.literal(SSL_PROVIDER_UNSPECIFIED_KEY_TYPE)),
        eabKid: z.string().trim(),
        eabHmacKey: z.string().trim(),
        availableInProjects: z.boolean(),
        default: z.boolean(),
    })
    .superRefine((value, ctx) => {
        if (value.kind === ESslProviderKind.LetsEncrypt) {
            return;
        }

        if (!value.eabKid) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["eabKid"],
                message: "EAB KID is required",
            });
        }

        if (!value.eabHmacKey) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["eabHmacKey"],
                message: "EAB HMAC is required",
            });
        }
    });

export type CreateOrEditSslProviderFormInput = z.input<typeof CreateOrEditSslProviderFormSchema>;
export type CreateOrEditSslProviderFormOutput = z.output<typeof CreateOrEditSslProviderFormSchema>;
