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
                    appKey?: string;
                    onSubmit?: (mount: AppStorageMount) => Promise<void>;
                    onClose?: () => void;
                    onError?: (error: Error) => void;
                    readOnly?: boolean;
                }) => {
                    const { onSubmit, onClose, onError, readOnly, projectKey, appKey } = options ?? {};
                    state.open({
                        projectKey,
                        appKey,
                        props: { onSubmit, onClose, onError, readOnly },
                    });
                },
                openEdit: (
                    mount: AppStorageMount & { _id: string },
                    options?: {
                        projectKey?: string;
                        appKey?: string;
                        onSubmit?: (mount: AppStorageMount) => Promise<void>;
                        onClose?: () => void;
                        onError?: (error: Error) => void;
                        readOnly?: boolean;
                    },
                ) => {
                    const { onSubmit, onClose, onError, readOnly, projectKey, appKey } = options ?? {};
                    state.openEdit(mount, {
                        projectKey,
                        appKey,
                        props: { onSubmit, onClose, onError, readOnly },
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
