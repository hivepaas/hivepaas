import { z } from "zod";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const AppPlacementSettingsEntitySchema = SettingsBaseEntitySchema.extend({
    excludeManagerNodes: z.boolean().optional(),
    excludeBuildNodes: z.boolean().optional(),
});
