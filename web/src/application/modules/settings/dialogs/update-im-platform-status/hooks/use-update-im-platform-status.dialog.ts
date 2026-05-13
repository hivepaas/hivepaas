import type { SettingImService } from "~/settings/domain";
import type { ImPlatformTableScope } from "~/settings/module-shared/components";

import type { UpdateImPlatformStatusDialogOptions } from "../types";

import { useUpdateImPlatformStatusDialogState } from "./use-update-im-platform-status.dialog.state";

function createHook() {
    return function useUpdateImPlatformStatusDialog(props: UpdateImPlatformStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateImPlatformStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: ImPlatformTableScope, imPlatform: SettingImService) => {
                    actions.open(scope, imPlatform, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateImPlatformStatusDialog = createHook();
