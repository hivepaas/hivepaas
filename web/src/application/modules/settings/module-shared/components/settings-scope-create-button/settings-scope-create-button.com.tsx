import type { ComponentProps, PropsWithChildren } from "react";

import { Button } from "@/components/ui";

import { SettingsScopePermissionAction } from "../settings-scope-permission-action";

interface SettingsScope {
    type: "settings" | "project";
}

interface Props extends PropsWithChildren {
    scope: SettingsScope;
    onClick: () => void;
    variant?: ComponentProps<typeof Button>["variant"];
}

export function SettingsScopeCreateButton({ scope, onClick, variant, children }: Props) {
    return (
        <SettingsScopePermissionAction
            scope={scope}
            action="write"
        >
            {({ isDenied }) => (
                <Button
                    variant={variant}
                    onClick={onClick}
                    disabled={isDenied}
                >
                    {children}
                </Button>
            )}
        </SettingsScopePermissionAction>
    );
}
