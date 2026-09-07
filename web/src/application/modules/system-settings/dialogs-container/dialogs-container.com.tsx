import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    ChangeKekDialog,
    QuickInstallSslCertDialog,
    RestartHivePaaSDialog,
    RestartTraefikDialog,
    RoutingChangeConfirmDialog,
    useChangeKekDialogState,
    useQuickInstallSslCertDialogState,
    useRestartHivePaaSDialogState,
    useRestartTraefikDialogState,
    useRoutingChangeConfirmDialogState,
} from "~/system-settings/dialogs";

function View() {
    const location = useLocation();
    const quickInstallSslCertDialog = useQuickInstallSslCertDialogState();
    const changeKekDialog = useChangeKekDialogState();
    const restartHivePaaSDialog = useRestartHivePaaSDialogState();
    const restartTraefikDialog = useRestartTraefikDialogState();
    const routingChangeConfirmDialog = useRoutingChangeConfirmDialogState();

    useUpdateEffect(() => {
        quickInstallSslCertDialog.destroy();
        changeKekDialog.destroy();
        restartHivePaaSDialog.destroy();
        restartTraefikDialog.destroy();
        // A no-op by design - see the store. A routing change stays on trial while
        // the operator browses elsewhere, and the dialog is the only thing that
        // can keep it.
        routingChangeConfirmDialog.destroy();
    }, [location]);

    return (
        <>
            <QuickInstallSslCertDialog />
            <ChangeKekDialog />
            <RestartHivePaaSDialog />
            <RestartTraefikDialog />
            <RoutingChangeConfirmDialog />
        </>
    );
}

export const SystemSettingsDialogsContainer = memo(View);
