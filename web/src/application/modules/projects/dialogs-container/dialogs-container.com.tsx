import React from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateOrEditAppConfigFileDialog,
    useCreateOrEditAppConfigFileDialogState,
} from "~/projects/dialogs/create-or-edit-app-config-file";
import {
    CreateOrEditAppSecretDialog,
    useCreateOrEditAppSecretDialogState,
} from "~/projects/dialogs/create-or-edit-app-secret";
import {
    CreateOrEditProjectSecretDialog,
    useCreateOrEditProjectSecretDialogState,
} from "~/projects/dialogs/create-or-edit-project-secret";
import { CreateProjectDialog, useCreateProjectDialogState } from "~/projects/dialogs/create-project";
import { CreateProjectAppDialog, useCreateProjectAppDialogState } from "~/projects/dialogs/create-project-app";
import {
    QuickInstallSslCertDialog,
    useQuickInstallSslCertDialogState,
} from "~/projects/dialogs/quick-install-ssl-cert";
import { StorageMountDialog, useStorageMountDialogState } from "~/projects/dialogs/storage-mount";
import {
    CreateOrEditBasicAuthDialog,
    UpdateBasicAuthStatusDialog,
    useCreateOrEditBasicAuthDialogState,
    useUpdateBasicAuthStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const createProjectDialog = useCreateProjectDialogState();
    const createProjectAppDialog = useCreateProjectAppDialogState();
    const createOrEditProjectSecretDialog = useCreateOrEditProjectSecretDialogState();
    const createOrEditAppConfigFileDialog = useCreateOrEditAppConfigFileDialogState();
    const createOrEditAppSecretDialog = useCreateOrEditAppSecretDialogState();
    const quickInstallSslCertDialog = useQuickInstallSslCertDialogState();
    const storageMountDialog = useStorageMountDialogState();
    const createOrEditBasicAuthDialog = useCreateOrEditBasicAuthDialogState();
    const updateBasicAuthStatusDialog = useUpdateBasicAuthStatusDialogState();

    useUpdateEffect(() => {
        createProjectDialog.destroy();
        createProjectAppDialog.destroy();
        createOrEditProjectSecretDialog.destroy();
        createOrEditAppConfigFileDialog.destroy();
        createOrEditAppSecretDialog.destroy();
        quickInstallSslCertDialog.destroy();
        storageMountDialog.destroy();
        createOrEditBasicAuthDialog.destroy();
        updateBasicAuthStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateProjectDialog />
            <CreateProjectAppDialog />
            <CreateOrEditProjectSecretDialog />
            <CreateOrEditAppConfigFileDialog />
            <CreateOrEditAppSecretDialog />
            <QuickInstallSslCertDialog />
            <StorageMountDialog />
            <CreateOrEditBasicAuthDialog />
            <UpdateBasicAuthStatusDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const ProjectsDialogsContainer = React.memo(View);
