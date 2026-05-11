import type { AppConfigFile } from "~/projects/domain";

import type { CreateOrEditAppConfigFileDialogOptions } from "../types";

import { useCreateOrEditAppConfigFileDialogState } from "./use-create-or-edit-app-config-file.dialog.state";

function createHook() {
    return function useCreateOrEditAppConfigFileDialog(props: CreateOrEditAppConfigFileDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditAppConfigFileDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, appId: string) => {
                    actions.open(projectId, appId, { props });
                },
                openEdit: (projectId: string, appId: string, configFile: AppConfigFile) => {
                    actions.openEdit(projectId, appId, configFile, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditAppConfigFileDialog = createHook();
