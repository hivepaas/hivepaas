import { create } from "zustand";
import type { AppSecret, ProjectSecret } from "~/projects/domain";

import type {
    CreateOrEditProjectSecretDialogOptions,
    CreateOrEditProjectSecretDialogScope,
    CreateOrEditProjectSecretDialogState,
} from "../types";

type State = CreateOrEditProjectSecretDialogState & CreateOrEditProjectSecretDialogOptions;

interface Actions {
    open: (
        projectId: string,
        scope: CreateOrEditProjectSecretDialogScope,
        appId?: string,
        options?: CreateOrEditProjectSecretDialogOptions,
    ) => void;
    openEdit: (
        projectId: string,
        scope: CreateOrEditProjectSecretDialogScope,
        secret: ProjectSecret | AppSecret,
        appId?: string,
        options?: CreateOrEditProjectSecretDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditProjectSecretDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        appId: null,
        scope: null,
    },

    props: {},

    open: (projectId, scope, appId, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                scope,
                appId,
            },
            ...options,
        });
    },

    openEdit: (projectId, scope, secret, appId, options = {}) => {
        set({
            state: {
                mode: "edit",
                projectId,
                scope,
                secret,
                appId,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
                projectId: null,
                appId: null,
                scope: null,
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
                    appId: null,
                    scope: null,
                },
                props: {},
            };
        });
    },
}));
