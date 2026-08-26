import type { UpdateProjectCommandPipeStatusDialogOptions } from "../types";

import { useUpdateProjectCommandPipeStatusDialogState } from "./use-update-project-command-pipe-status.dialog.state";

function createHook() {
    return function useUpdateProjectCommandPipeStatusDialog(
        props: UpdateProjectCommandPipeStatusDialogOptions["props"] = {},
    ) {
        const { state, props: _, ...actions } = useUpdateProjectCommandPipeStatusDialogState();

        return {
            state,
            actions: {
                open: (
                    projectId: string,
                    id: string,
                    options: UpdateProjectCommandPipeStatusDialogOptions & { env?: string } = {},
                ) => {
                    const { env, ...restOptions } = options;
                    actions.open(projectId, id, {
                        ...restOptions,
                        env,
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

export const useUpdateProjectCommandPipeStatusDialog = createHook();
