import { useState } from "react";

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { useUpdateEffect } from "react-use";

import { ProfileCommands } from "@application/shared/data/commands";

import { F2aSetupForm } from "../form";
import { useF2aSetupDialogState } from "../hooks";

type State = { QRCode: string; totpToken: string } | null;

export function F2aSetupDialog() {
    const [stateData, setStateData] = useState<State>(null);

    const {
        state,
        // props: { onSuccess = fnPlaceholder, onClose = fnPlaceholder, onError = fnPlaceholder } = {},
        // ...actions
    } = useF2aSetupDialogState();

    const { mutate: getProfile2FASetup, isPending: isGetProfile2FASetupPending } =
        ProfileCommands.useGetProfile2FASetup();

    useUpdateEffect(() => {
        if (state.mode === "open") {
            getProfile2FASetup(undefined, {
                onSuccess: data => {
                    setStateData({
                        QRCode: data.data.totpQRCode,
                        totpToken: data.data.totpToken,
                    });
                },
            });
        }
    }, [state.mode]);

    console.log("state", state);

    return (
        <Dialog open={state.mode === "open"}>
            <DialogHeader>
                <DialogTitle />
                <DialogDescription />
            </DialogHeader>
            <DialogContent>
                {stateData && (
                    <F2aSetupForm
                        isPending={isGetProfile2FASetupPending}
                        onSubmit={() => {}}
                        qrCode={stateData.QRCode}
                        totpToken={stateData.totpToken}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
