import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    JoinNewNodeDialog,
    UpdateNetworkStatusDialog,
    UpdateVolumeStatusDialog,
    useJoinNewNodeDialogState,
    useUpdateNetworkStatusDialogState,
    useUpdateVolumeStatusDialogState,
} from "../dialogs";

function View() {
    const location = useLocation();
    const joinNewNodeDialog = useJoinNewNodeDialogState();
    const updateNetworkStatusDialog = useUpdateNetworkStatusDialogState();
    const updateVolumeStatusDialog = useUpdateVolumeStatusDialogState();

    useUpdateEffect(() => {
        joinNewNodeDialog.destroy();
        updateNetworkStatusDialog.destroy();
        updateVolumeStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <JoinNewNodeDialog />
            <UpdateNetworkStatusDialog />
            <UpdateVolumeStatusDialog />
        </>
    );
}

export const ClusterDialogsContainer = React.memo(View);
