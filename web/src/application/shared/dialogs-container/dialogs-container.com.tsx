import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import { F2aSetupDialog, useF2aSetupDialogState } from "@application/shared/dialogs";

function View() {
    const location = useLocation();
    const f2aSetupDialog = useF2aSetupDialogState();

    useUpdateEffect(() => {
        f2aSetupDialog.destroy();
    }, [location]);

    return (
        <>
            <F2aSetupDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const CommonDialogsContainer = React.memo(View);
