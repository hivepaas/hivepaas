import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    QuickInstallSslCertDialog,
    useQuickInstallSslCertDialogState,
} from "~/system-settings/dialogs/quick-install-ssl-cert";

function View() {
    const location = useLocation();
    const quickInstallSslCertDialog = useQuickInstallSslCertDialogState();

    useUpdateEffect(() => {
        quickInstallSslCertDialog.destroy();
    }, [location]);

    return <QuickInstallSslCertDialog />;
}

export const SystemSettingsDialogsContainer = memo(View);
