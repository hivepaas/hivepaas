import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    ProvisionGithubAppDialog,
    UpdateGithubAppStatusDialog,
    UpdateRepoWebhookStatusDialog,
    useProvisionGithubAppDialogState,
    useUpdateGithubAppStatusDialogState,
    useUpdateRepoWebhookStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const provisionGithubAppDialog = useProvisionGithubAppDialogState();
    const updateGithubAppStatusDialog = useUpdateGithubAppStatusDialogState();
    const updateRepoWebhookStatusDialog = useUpdateRepoWebhookStatusDialogState();

    useUpdateEffect(() => {
        provisionGithubAppDialog.destroy();
        updateGithubAppStatusDialog.destroy();
        updateRepoWebhookStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <ProvisionGithubAppDialog />
            <UpdateGithubAppStatusDialog />
            <UpdateRepoWebhookStatusDialog />
        </>
    );
}

export const SourcesDialogsContainer = memo(View);
