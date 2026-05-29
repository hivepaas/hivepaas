import type { PropsWithChildren } from "react";

import { Button } from "@/components/ui";
import type { ModuleAction } from "@application/shared/permissions";

import { PopConfirm } from "@application/shared/components";

import { useSettingsScopePermissions } from "../../hooks";
import { SettingsScopePermissionAction } from "../settings-scope-permission-action";

interface SettingsScope {
    type: "settings" | "project";
}

interface SettingsScopeMenuButtonProps extends PropsWithChildren {
    scope: SettingsScope;
    action: ModuleAction;
    onClick?: () => void;
    isLoading?: boolean;
}

export function SettingsScopeMenuButton({
    scope,
    action,
    onClick,
    isLoading = false,
    children,
}: SettingsScopeMenuButtonProps) {
    return (
        <SettingsScopePermissionAction
            scope={scope}
            action={action}
            triggerClassName="w-full"
        >
            {({ isDenied }) => (
                <Button
                    className="justify-start py-1.5 w-full"
                    variant="ghost"
                    onClick={onClick}
                    disabled={isDenied}
                    isLoading={isLoading}
                >
                    {children}
                </Button>
            )}
        </SettingsScopePermissionAction>
    );
}

interface SettingsScopePopConfirmButtonProps extends PropsWithChildren {
    scope: SettingsScope;
    action: ModuleAction;
    title: string;
    confirmText: string;
    cancelText: string;
    description: string;
    onConfirm: () => void;
    isLoading?: boolean;
}

export function SettingsScopePopConfirmButton({
    scope,
    action,
    title,
    confirmText,
    cancelText,
    description,
    onConfirm,
    isLoading = false,
    children,
}: SettingsScopePopConfirmButtonProps) {
    const permission = useSettingsScopePermissions(scope);

    if (!permission.hasAccess(action)) {
        return (
            <SettingsScopeMenuButton
                scope={scope}
                action={action}
            >
                {children}
            </SettingsScopeMenuButton>
        );
    }

    return (
        <PopConfirm
            title={title}
            variant="destructive"
            confirmText={confirmText}
            cancelText={cancelText}
            description={description}
            onConfirm={onConfirm}
        >
            <Button
                className="justify-start py-1.5"
                variant="ghost"
                isLoading={isLoading}
            >
                {children}
            </Button>
        </PopConfirm>
    );
}
