import React from "react";

import { useLocation, useUpdateEffect } from "react-use";
import { CreateProjectDialog, useCreateProjectDialogState } from "~/projects/dialogs/create-project";
import { CreateProjectAppDialog, useCreateProjectAppDialogState } from "~/projects/dialogs/create-project-app";
import {
    CreateOrEditProjectSecretDialog,
    useCreateOrEditProjectSecretDialogState,
} from "~/projects/dialogs/create-or-edit-project-secret";

function View() {
    const location = useLocation();
    const createProjectDialog = useCreateProjectDialogState();
    const createProjectAppDialog = useCreateProjectAppDialogState();
    const createOrEditProjectSecretDialog = useCreateOrEditProjectSecretDialogState();

    useUpdateEffect(() => {
        createProjectDialog.destroy();
        createProjectAppDialog.destroy();
        createOrEditProjectSecretDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateProjectDialog />
            <CreateProjectAppDialog />
            <CreateOrEditProjectSecretDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const ProjectsDialogsContainer = React.memo(View);
