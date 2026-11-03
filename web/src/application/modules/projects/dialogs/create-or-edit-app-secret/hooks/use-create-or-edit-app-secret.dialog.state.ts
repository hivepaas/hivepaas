import { create } from "zustand";
import type { AppSecret } from "~/projects/domain";

import type { CreateOrEditAppSecretDialogOptions, CreateOrEditAppSecretDialogState } from "../types";

type State = CreateOrEditAppSecretDialogState & CreateOrEditAppSecretDialogOptions;

interface Actions {
    open: (projectId: string, appId: string, options?: CreateOrEditAppSecretDialogOptions) => void;
    openEdit: (
        projectId: string,
        appId: string,
        secret: AppSecret,
        options?: CreateOrEditAppSecretDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditAppSecretDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        appId: null,
    },

    props: {},

    open: (projectId, appId, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                appId,
            },
            ...options,
        });
    },

    openEdit: (projectId, appId, secret, options = {}) => {
        set({
            state: {
                mode: "edit",
                projectId,
                appId,
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
                appId: null,
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
                },
                props: {},
            };
        });
    },
}));
