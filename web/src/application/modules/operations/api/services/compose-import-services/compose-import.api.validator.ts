import { type AxiosResponse } from "axios";
import { z } from "zod";
import type { ComposePortAs, ComposeVolumeKind } from "~/operations/domain";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import { SpecImportPlanSchema } from "../spec-import-services/spec-import.api.validator";

import type { ComposeImport_Apply_Res, ComposeImport_Validate_Res } from "./compose-import.api.contracts";

const list = <T extends z.ZodTypeAny>(item: T) =>
    z
        .array(item)
        .nullish()
        .transform(value => value ?? []);

const PortSchema = z.object({
    published: z.number(),
    target: z.number(),
    protocol: z.string(),
    as: z.string().transform(value => value as ComposePortAs),
    default: z.string().transform(value => value as ComposePortAs),
    domain: z.string(),
    suggested: z.string(),
});

const VolumeSchema = z.object({
    target: z.string(),
    source: z.string(),
    kind: z.string().transform(value => value as ComposeVolumeKind),
    readOnly: z.boolean(),
    owner: z.string(),
});

const ServiceSchema = z.object({
    name: z.string(),
    app: z.string(),
    image: z.string(),
    build: z.boolean(),
    skipped: z.boolean(),
    reason: z.string(),
    mode: z.string(),
    replicas: z.number(),
    ports: list(PortSchema),
    volumes: list(VolumeSchema),
    aliases: list(z.string()),
    secrets: list(z.string()),
    dropped: list(z.string()),
    existing: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    useExisting: z
        .boolean()
        .nullish()
        .transform(value => value ?? false),
});

const ReviewSchema = z.object({
    project: z.object({
        id: z
            .string()
            .nullish()
            .transform(value => value ?? ""),
        name: z.string(),
        key: z.string(),
        env: z.string(),
        envKey: z.string(),
        newEnv: z
            .boolean()
            .nullish()
            .transform(value => value ?? false),
        fileName: z.string(),
    }),
    services: list(ServiceSchema),
    variables: list(
        z.object({
            name: z.string(),
            default: z.string(),
            required: z.boolean(),
            given: z.boolean(),
            secret: z.boolean(),
        }),
    ),
    needs: list(z.object({ path: z.string(), as: z.string(), by: list(z.string()), given: z.boolean() })),
    profiles: list(z.string()),
    plan: SpecImportPlanSchema.nullish().transform(value => value ?? undefined),
});

const ValidateSchema = z.object({
    data: ReviewSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const ApplySchema = z.object({
    data: z.object({
        project: z.object({ id: z.string() }).nullish(),
        plan: SpecImportPlanSchema,
        deployments: list(z.object({ appId: z.string(), deploymentId: z.string() })),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class ComposeImportApiValidator {
    validate = (response: AxiosResponse): ComposeImport_Validate_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ValidateSchema });

        return { data, meta };
    };

    apply = (response: AxiosResponse): ComposeImport_Apply_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ApplySchema });

        return {
            data: {
                plan: data.plan,
                deployments: data.deployments,
                projectId: data.project?.id,
                warning: meta?.warning === "" ? undefined : meta?.warning,
            },
            meta,
        };
    };
}
