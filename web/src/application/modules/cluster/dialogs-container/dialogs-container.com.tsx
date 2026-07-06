import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    JoinNewNodeDialog,
    UpdateVolumeStatusDialog,
    useJoinNewNodeDialogState,
    useUpdateVolumeStatusDialogState,
} from "../dialogs";

function View() {
    const location = useLocation();
    const joinNewNodeDialog = useJoinNewNodeDialogState();
    const updateVolumeStatusDialog = useUpdateVolumeStatusDialogState();

    useUpdateEffect(() => {
        joinNewNodeDialog.destroy();
        updateVolumeStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <JoinNewNodeDialog />
            <UpdateVolumeStatusDialog />
        </>
    );
}

export const ClusterDialogsContainer = React.memo(View);
