import { useState } from "react";

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { useUpdateEffect } from "react-use";
import { toast } from "sonner";

import { useProfileContext } from "@application/shared/context";
import { ProfileCommands } from "@application/shared/data/commands";
import { SessionQueries } from "@application/shared/data/queries";

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

    const { state, props: { isSetupRequired = false } = {}, ...actions } = useF2aSetupDialogState();

    const { mutate: complete2FASetup, isPending: isComplete2FASetupPending } = ProfileCommands.useComplete2FASetup();

    const { mutate: getProfile2FASetup, isPending: isGetProfile2FASetupPending } =
        ProfileCommands.useGetProfile2FASetup();

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
                    void refetch();
                    actions.close();
                    setStateData(null);
                },
            },
        );
    }

    const showCurrentPasscodeForm = state.mode === "change" && !stateData;
    const showSetupForm = stateData !== null;

    return (
        <Dialog
            open={state.mode !== "closed"}
            onOpenChange={open => {
                if (!open && isSetupRequired) {
                    return;
                }
                actions.close();
            }}
        >
            <DialogHeader>
                <DialogTitle />
                <DialogDescription />
            </DialogHeader>
            <DialogContent className="min-w-[400px] w-fit">
                {showCurrentPasscodeForm && (
                    <CurrentPasscodeForm
                        isPending={isGetProfile2FASetupPending}
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
            </DialogContent>
        </Dialog>
    );
}
