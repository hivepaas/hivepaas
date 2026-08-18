import type { AppPlacementSettings } from "~/settings/domain";

import type { SettingsAppPlacementFormSchemaInput } from "../schemas";

export function mapSettingsAppPlacementToFormInput(
    settings: AppPlacementSettings,
): SettingsAppPlacementFormSchemaInput {
    return {
        excludeManagerNodes: settings.excludeManagerNodes,
        excludeBuildNodes: settings.excludeBuildNodes,
    };
}
