import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import { JoinNewNodeDialog, useJoinNewNodeDialogState } from "../dialogs";

function View() {
    const location = useLocation();
    const joinNewNodeDialog = useJoinNewNodeDialogState();

    useUpdateEffect(() => {
        joinNewNodeDialog.destroy();
    }, [location]);

    return (
        <>
            <JoinNewNodeDialog />
        </>
    );
}

export const ClusterDialogsContainer = React.memo(View);
