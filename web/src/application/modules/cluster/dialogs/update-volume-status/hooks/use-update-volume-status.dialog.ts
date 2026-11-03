import type { VolumeManagementScope } from "~/cluster/module-shared/types";

import type { UpdateVolumeStatusDialogOptions } from "../types";

import { useUpdateVolumeStatusDialogState } from "./use-update-volume-status.dialog.state";

function createHook() {
    return function useUpdateVolumeStatusDialog(props: UpdateVolumeStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateVolumeStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: VolumeManagementScope, id: string, options: UpdateVolumeStatusDialogOptions = {}) => {
                    actions.open(scope, id, {
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

export const useUpdateVolumeStatusDialog = createHook();
