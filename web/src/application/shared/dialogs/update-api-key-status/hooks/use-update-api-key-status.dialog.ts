import type { ProfileApiKey } from "@application/shared/entities/profile";

import { type UpdateApiKeyStatusDialogOptions } from "../types/update-api-key-status.dialog.type";

import { useUpdateApiKeyStatusDialogState } from "./use-update-api-key-status.dialog.state";

function createHook() {
    return function useUpdateApiKeyStatusDialog(props: UpdateApiKeyStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateApiKeyStatusDialogState();

        return {
            state,
            actions: {
                open: (apiKey: ProfileApiKey) => {
                    actions.open(apiKey, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateApiKeyStatusDialog = createHook();

