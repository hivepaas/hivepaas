import { z } from "zod";

export const SettingsAppPlacementFormSchema = z.object({
    excludeManagerNodes: z.boolean(),
    excludeBuildNodes: z.boolean(),
});

export type SettingsAppPlacementFormSchemaInput = z.input<typeof SettingsAppPlacementFormSchema>;
export type SettingsAppPlacementFormSchemaOutput = z.output<typeof SettingsAppPlacementFormSchema>;

export const emptySettingsAppPlacementFormDefaults: SettingsAppPlacementFormSchemaInput = {
    excludeManagerNodes: false,
    excludeBuildNodes: false,
};
