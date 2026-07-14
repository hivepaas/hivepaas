import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    JoinNewNodeDialog,
    SetManagerNodesDialog,
    UpdateNetworkStatusDialog,
    UpdateVolumeStatusDialog,
    useJoinNewNodeDialogState,
    useSetManagerNodesDialogState,
    useUpdateNetworkStatusDialogState,
    useUpdateVolumeStatusDialogState,
} from "../dialogs";

function View() {
    const location = useLocation();
    const joinNewNodeDialog = useJoinNewNodeDialogState();
    const setManagerNodesDialog = useSetManagerNodesDialogState();
    const updateNetworkStatusDialog = useUpdateNetworkStatusDialogState();
    const updateVolumeStatusDialog = useUpdateVolumeStatusDialogState();

    useUpdateEffect(() => {
        joinNewNodeDialog.destroy();
        setManagerNodesDialog.destroy();
        updateNetworkStatusDialog.destroy();
        updateVolumeStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <JoinNewNodeDialog />
            <SetManagerNodesDialog />
            <UpdateNetworkStatusDialog />
            <UpdateVolumeStatusDialog />
        </>
    );
}

export const ClusterDialogsContainer = React.memo(View);
