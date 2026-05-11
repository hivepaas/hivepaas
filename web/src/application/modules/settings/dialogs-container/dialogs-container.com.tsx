import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateOrEditBasicAuthDialog,
    CreateOrEditRegistryAuthDialog,
    UpdateBasicAuthStatusDialog,
    UpdateRegistryAuthStatusDialog,
    useCreateOrEditBasicAuthDialogState,
    useCreateOrEditRegistryAuthDialogState,
    useUpdateBasicAuthStatusDialogState,
    useUpdateRegistryAuthStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const createOrEditBasicAuthDialog = useCreateOrEditBasicAuthDialogState();
    const updateBasicAuthStatusDialog = useUpdateBasicAuthStatusDialogState();
    const createOrEditRegistryAuthDialog = useCreateOrEditRegistryAuthDialogState();
    const updateRegistryAuthStatusDialog = useUpdateRegistryAuthStatusDialogState();

    useUpdateEffect(() => {
        createOrEditBasicAuthDialog.destroy();
        updateBasicAuthStatusDialog.destroy();
        createOrEditRegistryAuthDialog.destroy();
        updateRegistryAuthStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateOrEditBasicAuthDialog />
            <UpdateBasicAuthStatusDialog />
            <CreateOrEditRegistryAuthDialog />
            <UpdateRegistryAuthStatusDialog />
        </>
    );
}

export const SettingsDialogsContainer = memo(View);
