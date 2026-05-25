import type { CloudStorageTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditCloudStorageDialogOptions } from "../types";

import { useCreateOrEditCloudStorageDialogState } from "./use-create-or-edit-cloud-storage.dialog.state";

function createHook() {
    return function useCreateOrEditCloudStorageDialog(props: CreateOrEditCloudStorageDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditCloudStorageDialogState();

        return {
            state,
            actions: {
                open: (scope: CloudStorageTableScope, options: CreateOrEditCloudStorageDialogOptions = {}) => {
                    actions.open(scope, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                openEdit: (
                    scope: CloudStorageTableScope,
                    id: string,
                    options: CreateOrEditCloudStorageDialogOptions = {},
                ) => {
                    actions.openEdit(scope, id, {
                        ...options,
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

export const useCreateOrEditCloudStorageDialog = createHook();
