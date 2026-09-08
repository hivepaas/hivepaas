import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    ChangeKekDialog,
    QuickInstallSslCertDialog,
    RestartHivePaaSDialog,
    RestartTraefikDialog,
    SettingsChangeConfirmDialog,
    useChangeKekDialogState,
    useQuickInstallSslCertDialogState,
    useRestartHivePaaSDialogState,
    useRestartTraefikDialogState,
    useSettingsChangeConfirmDialogState,
} from "~/system-settings/dialogs";

function View() {
    const location = useLocation();
    const quickInstallSslCertDialog = useQuickInstallSslCertDialogState();
    const changeKekDialog = useChangeKekDialogState();
    const restartHivePaaSDialog = useRestartHivePaaSDialogState();
    const restartTraefikDialog = useRestartTraefikDialogState();
    const settingsChangeConfirmDialog = useSettingsChangeConfirmDialogState();

    useUpdateEffect(() => {
        quickInstallSslCertDialog.destroy();
        changeKekDialog.destroy();
        restartHivePaaSDialog.destroy();
        restartTraefikDialog.destroy();
        // A no-op by design - see the store. A change stays on trial while the
        // operator browses elsewhere, and the dialog is the only thing that can
        // keep it.
        settingsChangeConfirmDialog.destroy();
    }, [location]);

    return (
        <>
            <QuickInstallSslCertDialog />
            <ChangeKekDialog />
            <RestartHivePaaSDialog />
            <RestartTraefikDialog />
            <SettingsChangeConfirmDialog />
        </>
    );
}

export const SystemSettingsDialogsContainer = memo(View);
