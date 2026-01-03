import React from "react";

import { useLocation, useUpdateEffect } from "react-use";
import { CreateProjectDialog, useCreateProjectDialogState } from "~/projects/dialogs/create-project";

function View() {
    const location = useLocation();
    const createProjectDialog = useCreateProjectDialogState();

    useUpdateEffect(() => {
        createProjectDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateProjectDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const ProjectsDialogsContainer = React.memo(View);
