import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    CreateNetworkDialog,
    JoinNewNodeDialog,
    ViewNetworkDialog,
    useCreateNetworkDialogState,
    useJoinNewNodeDialogState,
    useViewNetworkDialogState,
} from "../dialogs";

function View() {
    const location = useLocation();
    const createNetworkDialog = useCreateNetworkDialogState();
    const joinNewNodeDialog = useJoinNewNodeDialogState();
    const viewNetworkDialog = useViewNetworkDialogState();

    useUpdateEffect(() => {
        createNetworkDialog.destroy();
        joinNewNodeDialog.destroy();
        viewNetworkDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateNetworkDialog />
            <JoinNewNodeDialog />
            <ViewNetworkDialog />
        </>
    );
}

export const ClusterDialogsContainer = React.memo(View);
