import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

function View() {
    const location = useLocation();

    useUpdateEffect(() => {
        // Cleanup dialogs on location change
    }, [location]);

    return null;
}

export const ClusterDialogsContainer = React.memo(View);
