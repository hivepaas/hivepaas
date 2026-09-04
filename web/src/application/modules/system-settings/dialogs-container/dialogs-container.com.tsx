import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    ChangeKekDialog,
    QuickInstallSslCertDialog,
    useChangeKekDialogState,
    useQuickInstallSslCertDialogState,
} from "~/system-settings/dialogs";

function View() {
    const location = useLocation();
    const quickInstallSslCertDialog = useQuickInstallSslCertDialogState();
    const changeKekDialog = useChangeKekDialogState();

    useUpdateEffect(() => {
        quickInstallSslCertDialog.destroy();
        changeKekDialog.destroy();
    }, [location]);

    return (
        <>
            <QuickInstallSslCertDialog />
            <ChangeKekDialog />
        </>
    );
}

export const SystemSettingsDialogsContainer = memo(View);
