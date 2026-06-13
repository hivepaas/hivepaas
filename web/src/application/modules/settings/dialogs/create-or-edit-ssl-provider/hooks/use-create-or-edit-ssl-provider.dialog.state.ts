import { create } from "zustand";
import type { SslProviderTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditSslProviderDialogOptions, CreateOrEditSslProviderDialogState } from "../types";

type State = CreateOrEditSslProviderDialogState & CreateOrEditSslProviderDialogOptions;

interface Actions {
    open: (scope: SslProviderTableScope, options?: CreateOrEditSslProviderDialogOptions) => void;
    openEdit: (scope: SslProviderTableScope, id: string, options?: CreateOrEditSslProviderDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditSslProviderDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (scope, options = {}) => {
        set({
            state: {
                mode: "open",
                scope,
            },
            ...options,
        });
    },

    openEdit: (scope, id, options = {}) => {
        set({
            state: {
                mode: "edit",
                scope,
                id,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
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
                },
                props: {},
            };
        });
    },
}));
