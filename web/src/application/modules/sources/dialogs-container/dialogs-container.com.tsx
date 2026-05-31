import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateOrEditGithubAppDialog,
    CreateOrEditRepoWebhookDialog,
    ProvisionGithubAppDialog,
    UpdateGithubAppStatusDialog,
    UpdateRepoWebhookStatusDialog,
    useCreateOrEditGithubAppDialogState,
    useCreateOrEditRepoWebhookDialogState,
    useProvisionGithubAppDialogState,
    useUpdateGithubAppStatusDialogState,
    useUpdateRepoWebhookStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const createOrEditGithubAppDialog = useCreateOrEditGithubAppDialogState();
    const createOrEditRepoWebhookDialog = useCreateOrEditRepoWebhookDialogState();
    const provisionGithubAppDialog = useProvisionGithubAppDialogState();
    const updateGithubAppStatusDialog = useUpdateGithubAppStatusDialogState();
    const updateRepoWebhookStatusDialog = useUpdateRepoWebhookStatusDialogState();

    useUpdateEffect(() => {
        createOrEditGithubAppDialog.destroy();
        createOrEditRepoWebhookDialog.destroy();
        provisionGithubAppDialog.destroy();
        updateGithubAppStatusDialog.destroy();
        updateRepoWebhookStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateOrEditGithubAppDialog />
            <CreateOrEditRepoWebhookDialog />
            <ProvisionGithubAppDialog />
            <UpdateGithubAppStatusDialog />
            <UpdateRepoWebhookStatusDialog />
        </>
    );
}

export const SourcesDialogsContainer = memo(View);
