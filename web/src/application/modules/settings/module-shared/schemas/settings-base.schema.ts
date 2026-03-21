import { z } from "zod";

import { ESettingStatus, ESettingType } from "@application/shared/enums";

export const SettingsBaseEntitySchema = z.object({
    id: z.string(),
    name: z.string(),
    description: z.string().optional().default(""),
    status: z.nativeEnum(ESettingStatus),
    type: z.nativeEnum(ESettingType),
    availableInProjects: z.boolean().optional(),
    default: z.boolean().optional(),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullish(),
    expireAt: z.coerce.date().nullish(),
});
