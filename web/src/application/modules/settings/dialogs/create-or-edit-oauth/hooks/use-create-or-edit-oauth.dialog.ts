import type { CreateOrEditOAuthDialogOptions } from "../types";

import { useCreateOrEditOAuthDialogState } from "./use-create-or-edit-oauth.dialog.state";

function createHook() {
    return function useCreateOrEditOAuthDialog(props: CreateOrEditOAuthDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditOAuthDialogState();

        return {
            state,
            actions: {
                open: () => {
                    actions.open({ props });
                },
                openEdit: (id: string) => {
                    actions.openEdit(id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditOAuthDialog = createHook();
