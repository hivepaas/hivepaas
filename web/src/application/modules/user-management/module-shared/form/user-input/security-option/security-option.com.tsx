import React from "react";

import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { type Path, useController, useFormContext } from "react-hook-form";

import { ESecuritySettings, EUserRole } from "@application/shared/enums";

const securityOptionMap: Record<ESecuritySettings, string> = {
    [ESecuritySettings.EnforceSSO]: "Enforce SSO",
    [ESecuritySettings.Password2FA]: "Password 2FA",
    [ESecuritySettings.PasswordOnly]: "Password Only",
};

/**
 * The options an admin account may use.
 *
 * Mirrors base.AdminSecurityOptions on the server, which refuses the rest: an
 * admin can grant themselves every module, project and capability, so a password
 * on its own is the whole of the defense around all of it. This is the same rule
 * said early, so the choice is never offered and then rejected on save - the
 * server stays the one that enforces it.
 */
const ADMIN_SECURITY_OPTIONS: ESecuritySettings[] = [ESecuritySettings.EnforceSSO, ESecuritySettings.Password2FA];

function isAllowedForRole(role: EUserRole | undefined, option: ESecuritySettings): boolean {
    if (role !== EUserRole.Admin) {
        return true;
    }
    return ADMIN_SECURITY_OPTIONS.includes(option);
}

function View<T>({ name, role, disabled = false }: Props<T>) {
    const { control } = useFormContext<Record<string, ESecuritySettings>>();

    const {
        field: securityOption,
        fieldState: { invalid },
    } = useController({
        control,
        name: name as string,
    });

    const { onChange } = securityOption;
    const selected = securityOption.value;
    const restricted = role === EUserRole.Admin;

    // Correct the selection when the operator promotes somebody to admin, and only
    // then. Switching the role while "Password Only" is selected would otherwise
    // leave a disabled tab selected and a save the server refuses; moving to 2FA is
    // the change they are going to have to make anyway, and it is visible - the tab
    // moves under them and the note below says why.
    //
    // Deliberately not on first render. An admin already on "Password Only" - the
    // bootstrap admin is seeded that way - would otherwise have their security
    // option quietly rewritten just by opening their page, and marked dirty, so the
    // next unrelated save would change it without anybody choosing to. The server
    // grandfathers those accounts for the same reason; this keeps the form honest
    // about it, showing the option they actually have while refusing to pick it
    // again once they leave it.
    const previousRole = React.useRef<EUserRole | undefined>(undefined);
    React.useEffect(() => {
        const roleChanged = previousRole.current !== undefined && previousRole.current !== role;
        previousRole.current = role;

        if (roleChanged && !isAllowedForRole(role, selected)) {
            onChange(ESecuritySettings.Password2FA);
        }
    }, [role, selected, onChange]);

    return (
        <div className="flex flex-col gap-1.5">
            <Tabs
                value={selected}
                onValueChange={value => {
                    if (disabled || !isAllowedForRole(role, value as ESecuritySettings)) {
                        return;
                    }

                    onChange(value as ESecuritySettings);
                }}
            >
                <TabsList>
                    {Object.entries(securityOptionMap).map(([value, label]) => (
                        <TabsTrigger
                            key={value}
                            value={value}
                            className="flex-1"
                            aria-invalid={invalid}
                            disabled={disabled || !isAllowedForRole(role, value as ESecuritySettings)}
                        >
                            {label}
                        </TabsTrigger>
                    ))}
                </TabsList>
            </Tabs>

            {restricted && (
                <p className="text-xs text-muted-foreground">
                    An admin can grant themselves anything, so a password on its own is not enough. Choose two-factor
                    authentication or enforced SSO.
                </p>
            )}
        </div>
    );
}

interface Props<T> {
    name: Path<T>;
    /** The role being assigned. Admins may not use a password on its own. */
    role?: EUserRole;
    disabled?: boolean;
}

export const SecurityOption = React.memo(View) as typeof View;
