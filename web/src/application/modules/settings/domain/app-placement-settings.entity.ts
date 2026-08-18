import type { SettingsBaseEntity } from "./settings.base.entity";

export interface AppPlacementSettings extends SettingsBaseEntity {
    excludeManagerNodes: boolean;
    excludeBuildNodes: boolean;
}
