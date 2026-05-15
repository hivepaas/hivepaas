import { z } from "zod";

import { EEmailKind, ESettingType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const EmailSMTPSchema = z.object({
    host: z.string(),
    port: z.number(),
    username: z.string(),
    displayName: z.string(),
    password: z.string(),
    ssl: z.boolean(),
});

const EmailHTTPFieldMappingSchema = z.object({
    fromAddress: z.string().optional().default(""),
    fromName: z.string().optional().default(""),
    toAddress: z.string().optional().default(""),
    toAddresses: z.string().optional().default(""),
    subject: z.string().optional().default(""),
    content: z.string().optional().default(""),
    password: z.string().optional().default(""),
});

const EmailHTTPSchema = z.object({
    endpoint: z.string(),
    method: z.enum(["GET", "POST", "PUT", "PATCH", "DELETE"]),
    contentType: z.string().optional().default(""),
    headers: z.record(z.string()).nullish(),
    fieldMapping: EmailHTTPFieldMappingSchema.nullish(),
    username: z.string().optional().default(""),
    displayName: z.string().optional().default(""),
    password: z.string().optional().default(""),
});

export const EmailSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.Email),
    kind: z.nativeEnum(EEmailKind),
    smtp: EmailSMTPSchema.nullish(),
    http: EmailHTTPSchema.nullish(),
    secretMasked: z.boolean().optional(),
    inherited: z.boolean().optional(),
});
