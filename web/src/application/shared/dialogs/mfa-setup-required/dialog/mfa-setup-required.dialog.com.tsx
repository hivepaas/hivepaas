import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";

import { useProfileContext } from "@application/shared/context";
import { SessionCommands } from "@application/shared/data/commands";
import { useF2aSetupDialog } from "@application/shared/dialogs/f2a-setup";

import { useAuthContext } from "@application/authentication/context";

import { Button } from "@/components/ui/button";

import { useMfaSetupRequiredDialogState } from "../hooks";

export function MfaSetupRequiredDialog() {
    const { state, props: { onActivate } = {}, ...actions } = useMfaSetupRequiredDialogState();
    const { clearProfile } = useProfileContext();
    const { clear: clearAuth } = useAuthContext();

    const f2aSetupDialog = useF2aSetupDialog({
        isSetupRequired: true,
    });

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

    function handleActivate() {
        actions.close();

        if (onActivate) {
            onActivate();
            return;
        }

        f2aSetupDialog.actions.open();
    }

    function handleDismissLogout() {
        actions.close();
        f2aSetupDialog.actions.close();
        clearProfile();
        clearAuth();
        logout();
    }

    return (
        <Dialog
            open={state.mode !== "closed"}
            onOpenChange={open => {
                if (!open) {
                    handleDismissLogout();
                }
            }}
        >
            <DialogFixedContent className="min-w-[400px] w-[480px] max-w-[calc(100vw-2rem)]">
                <DialogHeader>
                    <DialogTitle>Activate 2-Factor Authentication</DialogTitle>
                    <DialogDescription className="sr-only">
                        Your admin requires you to enable 2FA before continuing.
                    </DialogDescription>
                </DialogHeader>
                <DialogBody>
                    <p className="text-sm text-foreground">
                        Your admin requires you to enable 2FA to enhance security. Please complete this step to access
                        the system features.
                    </p>
                </DialogBody>
                <DialogActionFooter className="border-t-0">
                    <Button
                        type="button"
                        onClick={handleActivate}
                    >
                        Okay, Activate 2FA
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
