import type { SettingsBaseEntity } from "./settings.base.entity";

export interface AppPlacementSettings extends SettingsBaseEntity {
    excludeManagerNodes: boolean;
    excludeBuildNodes: boolean;
    /** `key=value` selectors; a bare key means `key=true`. */
    requireNodeLabels?: string[];
    excludeNodeLabels?: string[];
}
