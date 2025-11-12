import { type F2aSetupDialogOptions } from "../types";

import { useF2aSetupDialogState } from "./use-f2a-setup.dialog.state";

function createHook() {
    return function useF2aSetupDialog(props: F2aSetupDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useF2aSetupDialogState();

        return {
            state,
            actions: {
                open: () => {
                    actions.open({ props });
                },
                openChange: () => {
                    actions.openChange({ props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useF2aSetupDialog = createHook();
