import type { SettingImService } from "~/settings/domain";
import type { ImPlatformTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditImPlatformDialogOptions } from "../types";

import { useCreateOrEditImPlatformDialogState } from "./use-create-or-edit-im-platform.dialog.state";

function createHook() {
    return function useCreateOrEditImPlatformDialog(props: CreateOrEditImPlatformDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditImPlatformDialogState();

        return {
            state,
            actions: {
                open: (scope: ImPlatformTableScope) => {
                    actions.open(scope, { props });
                },
                openEdit: (scope: ImPlatformTableScope, imPlatform: SettingImService) => {
                    actions.openEdit(scope, imPlatform, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditImPlatformDialog = createHook();
