import { z } from "zod";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const AppPlacementSettingsEntitySchema = SettingsBaseEntitySchema.extend({
    excludeManagerNodes: z.boolean().optional(),
    excludeBuildNodes: z.boolean().optional(),
    requireNodeLabels: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    excludeNodeLabels: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
});
