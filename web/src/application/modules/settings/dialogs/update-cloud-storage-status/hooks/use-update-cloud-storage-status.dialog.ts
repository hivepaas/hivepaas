import type { CloudStorageTableScope } from "~/settings/module-shared/components";

import type { UpdateCloudStorageStatusDialogOptions } from "../types";

import { useUpdateCloudStorageStatusDialogState } from "./use-update-cloud-storage-status.dialog.state";

function createHook() {
    return function useUpdateCloudStorageStatusDialog(props: UpdateCloudStorageStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateCloudStorageStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: CloudStorageTableScope, id: string) => {
                    actions.open(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateCloudStorageStatusDialog = createHook();
