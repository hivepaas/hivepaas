import { type AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { HomeAttention_FindAll_Res } from "./home-attention.api.contracts";

const NamedObjectSchema = z.object({ id: z.string(), name: z.string().catch("") });

const AttentionItemSchema = z.object({
    kind: z.string(),
    severity: z.enum(["critical", "warning"]).catch("warning"),
    scope: z.string(),
    project: NamedObjectSchema.nullish().transform(value => value ?? null),
    env: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    app: NamedObjectSchema.nullish().transform(value => value ?? null),
    subject: z.string().catch(""),
    running: z
        .number()
        .nullish()
        .transform(value => value ?? 0),
    desired: z
        .number()
        .nullish()
        .transform(value => value ?? 0),
    restarts: z
        .number()
        .nullish()
        .transform(value => value ?? 0),
    lastError: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    nodeState: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    memoryLimitsBytes: z
        .number()
        .nullish()
        .transform(value => value ?? 0),
    memoryTotalBytes: z
        .number()
        .nullish()
        .transform(value => value ?? 0),
    since: z.coerce
        .date()
        .nullish()
        .transform(value => value ?? null),
    canAct: z.boolean().catch(false),
});

const FindAllSchema = z.object({
    data: z.object({ items: z.array(AttentionItemSchema).nullish() }),
    meta: BaseMetaApiSchema.nullish(),
});

export class HomeAttentionApiValidator {
    findAll = (response: AxiosResponse): HomeAttention_FindAll_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindAllSchema });
        return { data: data.items ?? [], meta };
    };
}
