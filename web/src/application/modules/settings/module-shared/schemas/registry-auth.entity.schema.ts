import { z } from "zod";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

/**
 * Registry-auth setting row from API (aligned with BE RegistryAuthResp + BaseSettingResp).
 */
export const RegistryAuthSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    description: z.string().optional(),
    type: z.string(),
    kind: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    inherited: z.boolean().optional(),
    address: z.string(),
    username: z.string(),
    password: z.string(),
    readonly: z.boolean(),
    secretMasked: z.boolean().optional(),
    ecr: z
        .object({
            region: z.string().nullish(),
            keyAuth: z.object({ id: z.string(), name: z.string().nullish() }).nullish(),
            roleArn: z.string().nullish(),
            tokenExpiresAt: z.coerce.date().nullish(),
        })
        .nullish()
        .transform(value =>
            value
                ? {
                      region: value.region ?? "",
                      keyAuth: value.keyAuth ? { id: value.keyAuth.id, name: value.keyAuth.name ?? "" } : null,
                      roleArn: value.roleArn ?? "",
                      tokenExpiresAt: value.tokenExpiresAt ?? null,
                  }
                : null,
        ),
});
