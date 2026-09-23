import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSRegistrySettings_CheckPush_Res,
    HivePaaSRegistrySettings_FindOne_Res,
    HivePaaSRegistrySettings_ProbeDomain_Res,
    HivePaaSRegistrySettings_RotateCredential_Res,
    HivePaaSRegistrySettings_UpdateOne_Res,
} from "./hivepaas-registry-settings.api.contracts";

const ObjectIdSchema = z.object({ id: z.string() }).nullish();

const StatusSchema = z
    .object({
        provisioned: z.boolean().catch(false),
        appId: z.string().optional(),
        reachable: z.boolean().catch(false),
        unreachable: z.string().optional(),
        repositories: z.number().catch(0),
        storedBytes: z.number().catch(0),
    })
    .nullish()
    .transform(value => value ?? { provisioned: false, reachable: false, repositories: 0, storedBytes: 0 });

const SettingsSchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    id: z.string().catch(""),
    name: z
        .string()
        .nullish()
        .transform(val => val ?? ""),
    status: z.string().catch("active"),
    type: z.string().catch("registry"),
    updateVer: z.number().catch(0),
    createdAt: z.coerce.date().catch(() => new Date()),
    enabled: z.boolean().catch(false),
    managed: z.boolean().catch(true),
    domain: z.string().catch(""),
    storage: z.object({
        type: z.enum(["volume", "s3"]).catch("volume"),
        volume: ObjectIdSchema,
        cloudStorage: ObjectIdSchema,
    }),
    cleanup: z.object({
        enabled: z.boolean().catch(true),
        mode: z.string().catch("policy"),
        keepLast: z.number().catch(10),
        keepDays: z.number().catch(30),
    }),
    dashboardEnabled: z.boolean().catch(false),
    memoryLimit: z.string().catch("1gb"),
    app: ObjectIdSchema,
    credential: ObjectIdSchema,
    registryStatus: StatusSchema,
    credentialRotation: z.object({ rotatedAt: z.string().optional(), graceEndsAt: z.string().optional() }).nullish(),
});

const FindOneSchema = z.object({
    data: SettingsSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const UpdateSchema = z.object({
    data: z
        .object({
            removedApp: z.boolean().catch(false),
            credentialKept: z.boolean().catch(false),
        })
        .nullish(),
    meta: BaseMetaApiSchema.nullish(),
});

const ProbeSchema = z.object({
    data: z.object({
        reached: z.boolean().catch(false),
        proxied: z.boolean().catch(false),
        evidence: z
            .array(z.string())
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const PushCheckSchema = z.object({
    data: z.object({
        ok: z.boolean().catch(false),
        statusCode: z.number().catch(0),
        detail: z.string().catch(""),
        elapsedMs: z.number().catch(0),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const RotateSchema = z.object({
    data: z.object({ graceEndsAt: z.string().catch("") }),
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSRegistrySettingsApiValidator {
    findOne = (response: AxiosResponse): HivePaaSRegistrySettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data: { settings: data, registryStatus: data.registryStatus }, meta };
    };

    updateOne = (response: AxiosResponse): HivePaaSRegistrySettings_UpdateOne_Res => {
        const { data } = parseApiResponse({ response, schema: UpdateSchema });
        return {
            data: {
                removedApp: data?.removedApp ?? false,
                credentialKept: data?.credentialKept ?? false,
            },
        };
    };

    probeDomain = (response: AxiosResponse): HivePaaSRegistrySettings_ProbeDomain_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ProbeSchema });
        return { data, meta };
    };

    checkPush = (response: AxiosResponse): HivePaaSRegistrySettings_CheckPush_Res => {
        const { data, meta } = parseApiResponse({ response, schema: PushCheckSchema });
        return { data, meta };
    };

    rotateCredential = (response: AxiosResponse): HivePaaSRegistrySettings_RotateCredential_Res => {
        const { data, meta } = parseApiResponse({ response, schema: RotateSchema });
        return { data, meta };
    };
}
