import type { AppStorageMount, ProjectStorageSettings } from "~/projects/domain";

import { useStorageMountDialogState } from "./use-storage-mount.dialog.state";

function createHook() {
    return function useStorageMountDialog() {
        const state = useStorageMountDialogState();

        return {
            state: state.state,
            actions: {
                open: (
                    projectRules?: ProjectStorageSettings,
                    options?: {
                        projectKey?: string;
                        appLocalKey?: string;
                        onSubmit?: (mount: AppStorageMount) => void;
                        onClose?: () => void;
                    },
                ) => {
                    const { onSubmit, onClose, projectKey, appLocalKey } = options ?? {};
                    state.open(projectRules, {
                        projectKey,
                        appLocalKey,
                        props: { onSubmit, onClose },
                    });
                },
                openEdit: (
                    mount: AppStorageMount & { _id: string },
                    projectRules?: ProjectStorageSettings,
                    options?: {
                        projectKey?: string;
                        appLocalKey?: string;
                        onSubmit?: (mount: AppStorageMount) => void;
                        onClose?: () => void;
                    },
                ) => {
                    const { onSubmit, onClose, projectKey, appLocalKey } = options ?? {};
                    state.openEdit(mount, projectRules, {
                        projectKey,
                        appLocalKey,
                        props: { onSubmit, onClose },
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
