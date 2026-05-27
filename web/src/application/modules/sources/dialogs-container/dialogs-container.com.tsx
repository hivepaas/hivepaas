import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateOrEditGithubAppDialog,
    ProvisionGithubAppDialog,
    UpdateGithubAppStatusDialog,
    useCreateOrEditGithubAppDialogState,
    useProvisionGithubAppDialogState,
    useUpdateGithubAppStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const createOrEditGithubAppDialog = useCreateOrEditGithubAppDialogState();
    const provisionGithubAppDialog = useProvisionGithubAppDialogState();
    const updateGithubAppStatusDialog = useUpdateGithubAppStatusDialogState();

    useUpdateEffect(() => {
        createOrEditGithubAppDialog.destroy();
        provisionGithubAppDialog.destroy();
        updateGithubAppStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateOrEditGithubAppDialog />
            <ProvisionGithubAppDialog />
            <UpdateGithubAppStatusDialog />
        </>
    );
}

export const SourcesDialogsContainer = memo(View);
