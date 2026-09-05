import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    ChangeKekDialog,
    QuickInstallSslCertDialog,
    RestartHivePaaSDialog,
    useChangeKekDialogState,
    useQuickInstallSslCertDialogState,
    useRestartHivePaaSDialogState,
} from "~/system-settings/dialogs";

function View() {
    const location = useLocation();
    const quickInstallSslCertDialog = useQuickInstallSslCertDialogState();
    const changeKekDialog = useChangeKekDialogState();
    const restartHivePaaSDialog = useRestartHivePaaSDialogState();

    useUpdateEffect(() => {
        quickInstallSslCertDialog.destroy();
        changeKekDialog.destroy();
        restartHivePaaSDialog.destroy();
    }, [location]);

    return (
        <>
            <QuickInstallSslCertDialog />
            <ChangeKekDialog />
            <RestartHivePaaSDialog />
        </>
    );
}

export const SystemSettingsDialogsContainer = memo(View);
