import type { ReactNode } from "react";

import { PermissionTooltipAction } from "@application/shared/permissions";
import type { ModuleAction, PermissionTooltipActionContext } from "@application/shared/permissions";

import { getSettingsScopeModuleId } from "../../hooks";

interface SettingsScope {
    type: "settings" | "project";
}

interface Props {
    scope: SettingsScope;
    action: ModuleAction;
    triggerClassName?: string;
    children: (context: PermissionTooltipActionContext<string>) => ReactNode;
}

export function SettingsScopePermissionAction({ scope, action, triggerClassName, children }: Props) {
    return (
        <PermissionTooltipAction
            id={getSettingsScopeModuleId(scope)}
            action={action}
            triggerClassName={triggerClassName}
        >
            {children}
        </PermissionTooltipAction>
    );
}
