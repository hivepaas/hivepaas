import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateOrEditBasicAuthDialog,
    UpdateBasicAuthStatusDialog,
    useCreateOrEditBasicAuthDialogState,
    useUpdateBasicAuthStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const createOrEditBasicAuthDialog = useCreateOrEditBasicAuthDialogState();
    const updateBasicAuthStatusDialog = useUpdateBasicAuthStatusDialogState();

    useUpdateEffect(() => {
        createOrEditBasicAuthDialog.destroy();
        updateBasicAuthStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateOrEditBasicAuthDialog />
            <UpdateBasicAuthStatusDialog />
        </>
    );
}

export const SettingsDialogsContainer = memo(View);
