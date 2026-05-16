import type { AppStorageMount } from "~/projects/domain";

import { useStorageMountDialogState } from "./use-storage-mount.dialog.state";

function createHook() {
    return function useStorageMountDialog() {
        const state = useStorageMountDialogState();

        return {
            state: state.state,
            actions: {
                open: (options?: {
                    projectKey?: string;
                    appLocalKey?: string;
                    onSubmit?: (mount: AppStorageMount) => Promise<void>;
                    onClose?: () => void;
                    onError?: (error: Error) => void;
                }) => {
                    const { onSubmit, onClose, onError, projectKey, appLocalKey } = options ?? {};
                    state.open({
                        projectKey,
                        appLocalKey,
                        props: { onSubmit, onClose, onError },
                    });
                },
                openEdit: (
                    mount: AppStorageMount & { _id: string },
                    options?: {
                        projectKey?: string;
                        appLocalKey?: string;
                        onSubmit?: (mount: AppStorageMount) => Promise<void>;
                        onClose?: () => void;
                        onError?: (error: Error) => void;
                    },
                ) => {
                    const { onSubmit, onClose, onError, projectKey, appLocalKey } = options ?? {};
                    state.openEdit(mount, {
                        projectKey,
                        appLocalKey,
                        props: { onSubmit, onClose, onError },
                    });
                },
                close: () => {
                    state.close();
                },
            },
        };
    };
}

export const useStorageMountDialog = createHook();
