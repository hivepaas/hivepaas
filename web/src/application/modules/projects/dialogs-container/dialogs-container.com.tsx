import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

function View() {
    const location = useLocation();

    useUpdateEffect(() => {
        // TODO: Destroy dialogs on location change
    }, [location]);

    return null;
}

export const ProjectsDialogsContainer = React.memo(View);
