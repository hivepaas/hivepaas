import type { ProjectSettingsImportKind } from "~/projects/data/commands";

import type { ImportProjectSettingsDialogOptions } from "../types";

import { useImportProjectSettingsDialogState } from "./use-import-project-settings.dialog.state";

function createHook() {
    return function useImportProjectSettingsDialog(props: ImportProjectSettingsDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useImportProjectSettingsDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, settingKind: ProjectSettingsImportKind) => {
                    actions.open(projectId, settingKind, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useImportProjectSettingsDialog = createHook();
