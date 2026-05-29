import { MODULE_IDS, type ResourceModuleId } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";
import type { UseConditionalModuleResult } from "@application/shared/permissions";

interface SettingsScope {
    type: "settings" | "project";
}

export function getSettingsScopeModuleId(scope: SettingsScope): ResourceModuleId {
    return scope.type === "project" ? MODULE_IDS.Project : MODULE_IDS.Settings;
}

export function useSettingsScopePermissions(scope: SettingsScope): UseConditionalModuleResult<ResourceModuleId> {
    return useConditionalModule({ id: getSettingsScopeModuleId(scope) });
}
