import { RefreshCw, Shield, ShieldOff } from "lucide-react";
import invariant from "tiny-invariant";
import type { UserBase } from "~/user-management/domain";

import { InfoBlock } from "@application/shared/components";
import { useProfileContext } from "@application/shared/context";
import { ESecuritySettings, EUserStatus } from "@application/shared/enums";

import { Button } from "@/components/ui/button";

interface Props {
    defaultValues: Pick<UserBase, "securityOption" | "status">;
}

export function Security({ defaultValues }: Props) {
    const { profile } = useProfileContext();
    invariant(profile, "Profile not found");

    // Logic: If Enforce SSO or status=pending, don't show Security section
    const shouldShowSecurity =
        defaultValues.securityOption !== ESecuritySettings.EnforceSSO && defaultValues.status !== EUserStatus.Pending;

    // Logic for MFA buttons
    const hasMfaSecret = profile.mfaSecret !== "";
    const isPassword2FA = defaultValues.securityOption === ESecuritySettings.Password2FA;
    const isPasswordOnly = defaultValues.securityOption === ESecuritySettings.PasswordOnly;

    if (!shouldShowSecurity) {
        return null;
    }

    return (
        <>
            <div className="h-px bg-border" />
            <InfoBlock title="Security">
                <div className="flex flex-col gap-4">
                    {/* Password Set Row */}
                    <div className="flex items-center flex-wrap gap-4">
                        <div className="w-[250px]">Password Set</div>
                        <Button
                            variant="outline"
                            size="default"
                            type="button"
                        >
                            <Shield className="size-4" />
                            Change Password
                        </Button>
                    </div>

                    {/* 2FA Row */}
                    {hasMfaSecret ? (
                        <div className="flex items-center flex-wrap gap-4">
                            <div className="w-[250px]">2-Factor Authentication Activated</div>
                            <div className="flex gap-2">
                                {isPassword2FA && (
                                    <Button
                                        variant="outline"
                                        size="default"
                                        type="button"
                                    >
                                        <RefreshCw className="size-4" />
                                        Reset 2FA
                                    </Button>
                                )}
                                {isPasswordOnly && (
                                    <Button
                                        variant="outline"
                                        size="default"
                                        type="button"
                                    >
                                        <ShieldOff className="size-4" />
                                        Deactivate 2FA
                                    </Button>
                                )}
                            </div>
                        </div>
                    ) : (
                        <div className="flex items-center flex-wrap gap-4">
                            <div className="w-[250px]">2-Factor Authentication Not Activated</div>
                            <Button
                                variant="outline"
                                size="default"
                                type="button"
                            >
                                <Shield className="size-4" />
                                Activate 2FA
                            </Button>
                        </div>
                    )}
                </div>
            </InfoBlock>
        </>
    );
}
