import type { UpdateProjectCommandTemplateStatusDialogOptions } from "../types";

import { useUpdateProjectCommandTemplateStatusDialogState } from "./use-update-project-command-template-status.dialog.state";

function createHook() {
    return function useUpdateProjectCommandTemplateStatusDialog(
        props: UpdateProjectCommandTemplateStatusDialogOptions["props"] = {},
    ) {
        const { state, props: _, ...actions } = useUpdateProjectCommandTemplateStatusDialogState();

        return {
            state,
            actions: {
                open: (
                    projectId: string,
                    id: string,
                    options: UpdateProjectCommandTemplateStatusDialogOptions = {},
                ) => {
                    actions.open(projectId, id, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateProjectCommandTemplateStatusDialog = createHook();
