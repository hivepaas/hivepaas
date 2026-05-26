import { create } from "zustand";

import type { ProjectUserAccessesDialogOptions, ProjectUserAccessesDialogState } from "../types";

type State = ProjectUserAccessesDialogState & ProjectUserAccessesDialogOptions;

interface Actions {
    open: (projectId: string, projectName: string, options?: ProjectUserAccessesDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useProjectUserAccessesDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        projectName: null,
    },

    props: {},

    open: (projectId, projectName, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                projectName,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
                projectId: null,
                projectName: null,
            },
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
                state: {
                    mode: "closed",
                    projectId: null,
                    projectName: null,
                },
                props: {},
            };
        });
    },
}));
