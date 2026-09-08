import type { AxiosResponse } from "axios";
import { z } from "zod";
import type { SettingsPendingChange } from "~/system-settings/domain";
import type { SettingsPendingChangeSchema } from "~/system-settings/module-shared/schemas";
import { SettingsPendingChangeSchema as PendingChangeSchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    TraefikConfigOptions_ConfirmChange_Res,
    TraefikConfigOptions_FindOne_Res,
    TraefikConfigOptions_RevertChange_Res,
    TraefikConfigOptions_UpdateOne_Res,
} from "./traefik-config-options.api.contracts";

const StartupCommandSchema = z.object({
    logLevel: z.string().optional().default(""),
    accessLog: z.boolean().optional().default(false),
    http3: z.boolean().optional().default(false),
    fastProxy: z.boolean().optional().default(false),
    openPorts: z.array(z.string()).optional().default([]),
    args: z.array(z.string()).default([]),
});

const FindOneSchema = z.object({
    data: z.object({
        startupCommand: StartupCommandSchema,
        pendingChange: PendingChangeSchema.nullish(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

const UpdateOneSchema = z.object({
    data: PendingChangeSchema.nullish(),
    meta: BaseMetaApiSchema.nullish(),
});

const RevertChangeSchema = z.object({
    data: z
        .object({
            reverted: z.boolean(),
            reason: z.string().nullish(),
        })
        .nullish(),
    meta: BaseMetaApiSchema.nullish(),
});

function mapPendingChange(
    raw: z.infer<typeof SettingsPendingChangeSchema> | null | undefined,
): SettingsPendingChange | null {
    if (raw == null) {
        return null;
    }
    return {
        changeId: raw.changeId,
        appliedAt: raw.appliedAt,
        confirmableFrom: raw.confirmableFrom,
        deadlineAt: raw.deadlineAt,
    };
}

export class TraefikConfigOptionsApiValidator {
    findOne = (response: AxiosResponse): TraefikConfigOptions_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                startupCommand: data.startupCommand,
                pendingChange: mapPendingChange(data.pendingChange ?? undefined),
            },
            meta,
        };
    };

    updateOne = (response: AxiosResponse): TraefikConfigOptions_UpdateOne_Res => {
        const { data } = parseApiResponse({ response, schema: UpdateOneSchema });
        return { data: { pendingChange: mapPendingChange(data ?? undefined) } };
    };

    confirmChange = (response: AxiosResponse): TraefikConfigOptions_ConfirmChange_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    revertChange = (response: AxiosResponse): TraefikConfigOptions_RevertChange_Res => {
        const { data } = parseApiResponse({ response, schema: RevertChangeSchema });
        return { data: { reverted: data?.reverted ?? false, reason: data?.reason ?? null } };
    };
}
