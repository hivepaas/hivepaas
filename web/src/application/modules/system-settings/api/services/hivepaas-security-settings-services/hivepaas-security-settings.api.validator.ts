import type { AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSSecuritySettings_FindOne_Res,
    HivePaaSSecuritySettings_UpdateOne_Res,
} from "./hivepaas-security-settings.api.contracts";

const FindOneSchema = z.object({
    data: z.object({
        returnSecretsViaApi: z.boolean(),
        alwaysReturnSecretTypes: z
            .array(z.string())
            .nullish()
            .transform(value => value ?? []),
        // Absent from a server older than the switch, which has it off.
        allowPrivilegedApps: z.boolean().optional().default(false),
        privilegedApps: z
            .array(
                z.object({
                    appId: z.string(),
                    appName: z.string(),
                    projectId: z.string(),
                    projectName: z.string(),
                    projectEnvKey: z.string(),
                    projectEnvName: z.string(),
                }),
            )
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSSecuritySettingsApiValidator {
    findOne = (response: AxiosResponse): HivePaaSSecuritySettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                returnSecretsViaApi: data.returnSecretsViaApi,
                alwaysReturnSecretTypes: data.alwaysReturnSecretTypes,
                allowPrivilegedApps: data.allowPrivilegedApps,
                privilegedApps: data.privilegedApps,
            },
            meta,
        };
    };

    updateOne = (response: AxiosResponse): HivePaaSSecuritySettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
