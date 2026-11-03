import { create } from "zustand";
import type { ProjectSettingsImportKind } from "~/projects/data/commands";

import type { ImportProjectSettingsDialogOptions, ImportProjectSettingsDialogState } from "../types";

type State = ImportProjectSettingsDialogState & ImportProjectSettingsDialogOptions;

interface Actions {
    open: (
        projectId: string,
        settingKind: ProjectSettingsImportKind,
        options?: ImportProjectSettingsDialogOptions & { env?: string },
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const closedState = {
    mode: "closed" as const,
    projectId: null,
    env: null,
    settingKind: null,
};

export const useImportProjectSettingsDialogState = create<State & Actions>()(set => ({
    state: closedState,

    props: {},

    open: (projectId, settingKind, options = {}) => {
        const { env, ...restOptions } = options;
        set({
            state: {
                mode: "open",
                projectId,
                env,
                settingKind,
            },
            ...restOptions,
        });
    },

    close: () => {
        set({
            state: closedState,
        });
    },

    clear: () => {
        set({
            props: {},
        });
    },

    destroy: () => {
        set(state => {
            if (state.state.mode === "closed") {
                return state;
            }

            return {
                state: closedState,
                props: {},
            };
        });
    },
}));
