import type { RunNowTaskCreatedDialogOptions } from "../types";

import { useRunNowTaskCreatedDialogState } from "./use-run-now-task-created.dialog.state";

function createHook() {
    return function useRunNowTaskCreatedDialog(props: RunNowTaskCreatedDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useRunNowTaskCreatedDialogState();

        return {
            state,
            actions: {
                open: (taskId: string) => {
                    actions.open(taskId, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useRunNowTaskCreatedDialog = createHook();
