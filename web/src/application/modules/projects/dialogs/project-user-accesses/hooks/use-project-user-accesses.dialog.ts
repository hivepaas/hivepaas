import type { ProjectEnvEntity } from "~/projects/domain";

import type { ProjectUserAccessesDialogOptions } from "../types";

import { useProjectUserAccessesDialogState } from "./use-project-user-accesses.dialog.state";

function createHook() {
    return function useProjectUserAccessesDialog(props: ProjectUserAccessesDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useProjectUserAccessesDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, projectName: string, envs: ProjectEnvEntity[]) => {
                    actions.open(projectId, projectName, envs, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useProjectUserAccessesDialog = createHook();
