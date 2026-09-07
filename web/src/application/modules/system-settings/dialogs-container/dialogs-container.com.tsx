import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    ChangeKekDialog,
    QuickInstallSslCertDialog,
    RestartHivePaaSDialog,
    RestartTraefikDialog,
    useChangeKekDialogState,
    useQuickInstallSslCertDialogState,
    useRestartHivePaaSDialogState,
    useRestartTraefikDialogState,
} from "~/system-settings/dialogs";

function View() {
    const location = useLocation();
    const quickInstallSslCertDialog = useQuickInstallSslCertDialogState();
    const changeKekDialog = useChangeKekDialogState();
    const restartHivePaaSDialog = useRestartHivePaaSDialogState();
    const restartTraefikDialog = useRestartTraefikDialogState();

    useUpdateEffect(() => {
        quickInstallSslCertDialog.destroy();
        changeKekDialog.destroy();
        restartHivePaaSDialog.destroy();
        restartTraefikDialog.destroy();
    }, [location]);

    return (
        <>
            <QuickInstallSslCertDialog />
            <ChangeKekDialog />
            <RestartHivePaaSDialog />
            <RestartTraefikDialog />
        </>
    );
}

export const SystemSettingsDialogsContainer = memo(View);
