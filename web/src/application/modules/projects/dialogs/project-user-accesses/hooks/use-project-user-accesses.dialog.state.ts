import { create } from "zustand";
import type { ProjectEnvEntity } from "~/projects/domain";

import type { ProjectUserAccessesDialogOptions, ProjectUserAccessesDialogState } from "../types";

type State = ProjectUserAccessesDialogState & ProjectUserAccessesDialogOptions;

interface Actions {
    open: (
        projectId: string,
        projectName: string,
        envs: ProjectEnvEntity[],
        options?: ProjectUserAccessesDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useProjectUserAccessesDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        projectName: null,
        envs: null,
    },

    props: {},

    open: (projectId, projectName, envs, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                projectName,
                envs,
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
                envs: null,
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
                    envs: null,
                },
                props: {},
            };
        });
    },
}));
