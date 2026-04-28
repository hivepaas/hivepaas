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
                        onSubmit?: (mount: AppStorageMount) => void;
                        onClose?: () => void;
                    },
                ) => {
                    state.open(projectRules, { props: options });
                },
                openEdit: (
                    mount: AppStorageMount & { _id: string },
                    projectRules?: ProjectStorageSettings,
                    options?: {
                        onSubmit?: (mount: AppStorageMount) => void;
                        onClose?: () => void;
                    },
                ) => {
                    state.openEdit(mount, projectRules, { props: options });
                },
                close: () => {
                    state.close();
                },
            },
        };
    };
}

export const useStorageMountDialog = createHook();
