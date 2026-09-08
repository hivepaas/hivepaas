import type { AxiosResponse } from "axios";
import { z } from "zod";

import type { SettingUsages_FindMany_Res } from "@application/shared/api/services";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Everything past type and id is optional on purpose: the server fills in what it
 * can, and a row it could not describe still has to arrive. A reference to
 * something already deleted is exactly what this endpoint exists to report.
 */
const SettingUsageSchema = z.object({
    type: z.string(),
    id: z.string(),
    name: z.string().nullish(),
    settingType: z.string().nullish(),
    scope: z.string().nullish(),
    projectId: z.string().nullish(),
    projectEnvKey: z.string().nullish(),
    appId: z.string().nullish(),
    appName: z.string().nullish(),
    userId: z.string().nullish(),
});

const FindManySchema = z.object({
    data: z.array(SettingUsageSchema).nullish(),
    meta: BaseMetaApiSchema.nullish(),
});

export class SettingUsageApiValidator {
    findMany = (response: AxiosResponse): SettingUsages_FindMany_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManySchema });

        return {
            data: (data ?? []).map(usage => ({
                type: usage.type,
                id: usage.id,
                name: usage.name ?? undefined,
                settingType: usage.settingType ?? undefined,
                scope: usage.scope ?? undefined,
                projectId: usage.projectId ?? undefined,
                projectEnvKey: usage.projectEnvKey ?? undefined,
                appId: usage.appId ?? undefined,
                appName: usage.appName ?? undefined,
                userId: usage.userId ?? undefined,
            })),
            meta,
        };
    };
}
