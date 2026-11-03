import { type GlobalAlertDialogOptions } from "../types/global-alert.dialog.type";

import { useGlobalAlertDialogState } from "./use-global-alert.dialog.state";

function createHook() {
    return function useGlobalAlertDialog() {
        const { mode, props, ...actions } = useGlobalAlertDialogState();

        return {
            state: {
                mode,
            },
            props,
            actions: {
                open: (options: GlobalAlertDialogOptions) => {
                    actions.open(options);
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useGlobalAlertDialog = createHook();
