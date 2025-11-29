import { type CreateProfileApiKeyDialogOptions } from "../types";

import { useCreateProfileApiKeyDialogState } from "./use-create-profile-api-key.dialog.state";

function createHook() {
    return function useCreateProfileApiKeyDialog(props: CreateProfileApiKeyDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateProfileApiKeyDialogState();

        return {
            state,
            actions: {
                open: () => {
                    actions.open({ props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateProfileApiKeyDialog = createHook();

