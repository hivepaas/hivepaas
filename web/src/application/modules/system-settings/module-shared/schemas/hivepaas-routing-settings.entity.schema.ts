import { z } from "zod";

import { SettingsPendingChangeSchema } from "./settings-probation.entity.schema";

const SettingRefSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .passthrough();

export const HivePaaSRoutingClientConfigSchema = z.object({
    enabled: z.boolean(),
    allowedIPs: z.array(z.string()).nullish(),
});

export const HivePaaSRoutingDomainSchema = z.object({
    enabled: z.boolean(),
    domain: z.string(),
    sslCert: SettingRefSchema.nullish(),
    clientConfig: HivePaaSRoutingClientConfigSchema.nullish(),
});

export const HivePaaSRoutingSettingsEntitySchema = z.object({
    domains: z.array(HivePaaSRoutingDomainSchema).nullish(),
    updateVer: z.number(),
    pendingChange: SettingsPendingChangeSchema.nullish(),
});

export const HivePaaSHttpSettingsEntitySchema = HivePaaSRoutingSettingsEntitySchema;
