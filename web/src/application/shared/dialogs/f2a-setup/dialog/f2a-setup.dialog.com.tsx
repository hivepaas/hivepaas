import { useState } from "react";

import { Dialog, DialogDescription, DialogFixedContent, DialogTitle } from "@components/ui/dialog";
import { useUpdateEffect } from "react-use";
import { toast } from "sonner";

import { useProfileContext } from "@application/shared/context";
import { ProfileCommands, SessionCommands } from "@application/shared/data/commands";
import { SessionQueries } from "@application/shared/data/queries";

import { useAuthContext } from "@application/authentication/context";

import {
    CurrentPasscodeForm,
    type CurrentPasscodeSchemaOutput,
    F2aSetupForm,
    type F2aSetupSchemaOutput,
} from "../form";
import { useF2aSetupDialogState } from "../hooks";

type State = { qrCode: string; totpToken: string; secretKey: string } | null;

export function F2aSetupDialog() {
    const [stateData, setStateData] = useState<State>(null);
    const { setProfile, clearProfile } = useProfileContext();
    const { clear: clearAuth } = useAuthContext();

    const { state, ...actions } = useF2aSetupDialogState();

    const { mutate: logout } = SessionCommands.useLogout({
        onSuccess: () => {
            clearProfile();
            clearAuth();
        },
        onSessionInvalid: () => {
            clearProfile();
            clearAuth();
        },
    });

    const { mutate: complete2FASetup, isPending: isComplete2FASetupPending } = ProfileCommands.useComplete2FASetup();

    const { mutate: getProfile2FASetup, isPending: isGetProfile2FASetupPending } =
        ProfileCommands.useGetProfile2FASetup();

    const { mutate: removeMfaTotp, isPending: isRemoveMfaTotpPending } = ProfileCommands.useRemoveMfaTotp();

    const { refetch } = SessionQueries.useGetProfile({
        onSuccess: ({ data }) => {
            setProfile(data);
        },
        onSessionInvalid: clearProfile,
        enabled: false,
    });

    useUpdateEffect(() => {
        if (state.mode === "open") {
            getProfile2FASetup(
                {},
                {
                    onSuccess: data => {
                        setStateData({
                            qrCode: data.data.totpQRCode,
                            secretKey: data.data.secretKey,
                            totpToken: data.data.totpToken,
                        });
                    },
                },
            );
        }
    }, [state.mode]);

    function onCurrentPasscodeSubmit(values: CurrentPasscodeSchemaOutput) {
        if (state.mode === "deactivate") {
            removeMfaTotp(
                {
                    passcode: (values as { currentPasscode: string }).currentPasscode,
                },
                {
                    onSuccess: () => {
                        toast.success("2FA deactivated successfully");
                        void refetch();
                        actions.close();
                        setStateData(null);
                    },
                },
            );
            return;
        }

        getProfile2FASetup(
            {
                passcode: (values as { currentPasscode: string }).currentPasscode,
            },
            {
                onSuccess: data => {
                    setStateData({
                        qrCode: data.data.totpQRCode,
                        secretKey: data.data.secretKey,
                        totpToken: data.data.totpToken,
                    });
                },
            },
        );
    }

    function onSubmit(values: F2aSetupSchemaOutput) {
        if (!stateData?.totpToken) {
            return;
        }

        complete2FASetup(
            {
                totpToken: stateData.totpToken,
                passcode: values.passcode,
            },
            {
                onSuccess: () => {
                    toast.success("2FA setup completed successfully");
                    clearAuth();
                    void refetch();
                    actions.close();
                    setStateData(null);
                },
            },
        );
    }

    const showCurrentPasscodeForm = (state.mode === "change" && !stateData) || state.mode === "deactivate";
    const showSetupForm = stateData !== null && state.mode !== "deactivate";

    function handleDismissLogout() {
        // Close first and clear MFA flag so ModuleLayout cannot reopen the intro
        // while logout is in flight. Clear profile only after logout removes the
        // token — otherwise ApplicationProfileInit refetches and re-enforces 2FA.
        actions.close();
        setStateData(null);
        clearAuth();
        logout();
    }

    return (
        <Dialog
            open={state.mode !== "closed"}
            onOpenChange={open => {
                if (!open) {
                    // Read latest auth so a successful setup (clearAuth then close)
                    // does not look like an enforced dismiss.
                    const authData = useAuthContext.getState().data;
                    const isEnforcedSetup = "mfaSetupRequired" in authData && authData.mfaSetupRequired;

                    if (isEnforcedSetup) {
                        handleDismissLogout();
                        return;
                    }

                    actions.close();
                    setStateData(null);
                }
            }}
        >
            <DialogFixedContent className="min-w-[400px] w-fit">
                <DialogTitle className="sr-only">Two-factor authentication</DialogTitle>
                <DialogDescription className="sr-only">
                    Set up, change, or deactivate two-factor authentication.
                </DialogDescription>
                {showCurrentPasscodeForm && (
                    <CurrentPasscodeForm
                        isPending={state.mode === "deactivate" ? isRemoveMfaTotpPending : isGetProfile2FASetupPending}
                        onSubmit={onCurrentPasscodeSubmit}
                    />
                )}
                {showSetupForm && (
                    <F2aSetupForm
                        isPending={isGetProfile2FASetupPending || isComplete2FASetupPending}
                        onSubmit={onSubmit}
                        qrCode={stateData.qrCode}
                        secretKey={stateData.secretKey}
                        totpToken={stateData.totpToken}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
