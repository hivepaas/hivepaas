import type { ProjectUserAccessesDialogOptions } from "../types";

import { useProjectUserAccessesDialogState } from "./use-project-user-accesses.dialog.state";

function createHook() {
    return function useProjectUserAccessesDialog(props: ProjectUserAccessesDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useProjectUserAccessesDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, projectName: string) => {
                    actions.open(projectId, projectName, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useProjectUserAccessesDialog = createHook();
