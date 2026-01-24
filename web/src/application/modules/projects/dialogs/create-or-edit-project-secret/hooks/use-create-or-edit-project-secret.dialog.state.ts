import { create } from "zustand";
import type { ProjectSecret } from "~/projects/domain";

import type { CreateOrEditProjectSecretDialogOptions, CreateOrEditProjectSecretDialogState } from "../types";

type State = CreateOrEditProjectSecretDialogState & CreateOrEditProjectSecretDialogOptions;

interface Actions {
    open: (projectId: string, options?: CreateOrEditProjectSecretDialogOptions) => void;
    openEdit: (projectId: string, secret: ProjectSecret, options?: CreateOrEditProjectSecretDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditProjectSecretDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
    },

    props: {},

    open: (projectId, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
            },
            ...options,
        });
    },

    openEdit: (projectId, secret, options = {}) => {
        set({
            state: {
                mode: "edit",
                projectId,
                secret,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
                projectId: null,
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
                },
                props: {},
            };
        });
    },
}));
