import { type AxiosResponse } from "axios";
import { z } from "zod";
import type {
    SpecImportIssue,
    SpecImportNodeAction,
    SpecImportNodeKind,
    SpecImportNodeOutcome,
    SpecImportSeverity,
    SpecSecretsMode,
} from "~/operations/domain";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { SpecImport_Apply_Res, SpecImport_Validate_Res } from "./spec-import.api.contracts";

const IssueSchema = z
    .object({
        severity: z
            .string()
            .optional()
            .transform(value => (value ? (value as SpecImportSeverity) : undefined)),
        code: z.string(),
        path: z.string(),
        detail: z.record(z.string(), z.unknown()).nullish(),
        availableIn: z.string().optional(),
        action: z.string().optional(),
        hint: z.string().optional(),
    })
    .transform(
        (issue): SpecImportIssue => ({
            ...issue,
            detail: issue.detail ?? undefined,
        }),
    );

const NodeSchema = z.object({
    path: z.string(),
    kind: z.string().transform(value => value as SpecImportNodeKind),
    key: z.string().optional(),
    name: z.string().optional(),
    selected: z.boolean(),
    selectedBy: z.string().optional(),
    action: z.string().transform(value => value as SpecImportNodeAction),
    matchedBy: z.string().optional(),
    changes: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    restart: z.boolean(),
    deploy: z.boolean(),
    issues: z
        .array(IssueSchema)
        .nullish()
        .transform(value => value ?? []),
    notes: z
        .array(IssueSchema)
        .nullish()
        .transform(value => value ?? []),
    outcome: z
        .string()
        .optional()
        .transform(value => (value ? (value as SpecImportNodeOutcome) : undefined)),
    error: z.string().optional(),
});

const PlanSchema = z.object({
    bundle: z.object({
        apiVersion: z.string(),
        scope: z.string(),
        exportedAt: z.coerce.date(),
        sourceAppVersion: z.string(),
        secretsMode: z.string().transform(value => value as SpecSecretsMode),
        digest: z.string(),
    }),
    nodes: z
        .array(NodeSchema)
        .nullish()
        .transform(value => value ?? []),
    summary: z
        .record(z.string(), z.number())
        .nullish()
        .transform(value => value ?? {}),
    planHash: z.string(),
});

const ValidateSchema = z.object({
    data: PlanSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const ApplySchema = z.object({
    data: z.object({
        plan: PlanSchema,
        deployments: z
            .array(z.object({ appId: z.string(), deploymentId: z.string() }))
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class SpecImportApiValidator {
    validate = (response: AxiosResponse): SpecImport_Validate_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ValidateSchema });

        return { data, meta };
    };

    apply = (response: AxiosResponse): SpecImport_Apply_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ApplySchema });

        return {
            data: { ...data, warning: meta?.warning === "" ? undefined : meta?.warning },
            meta,
        };
    };
}
