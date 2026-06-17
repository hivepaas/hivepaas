import { create } from "zustand";
import type { AcmeDnsProviderTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditAcmeDnsProviderDialogOptions, CreateOrEditAcmeDnsProviderDialogState } from "../types";

type State = CreateOrEditAcmeDnsProviderDialogState & CreateOrEditAcmeDnsProviderDialogOptions;

interface Actions {
    open: (scope: AcmeDnsProviderTableScope, options?: CreateOrEditAcmeDnsProviderDialogOptions) => void;
    openEdit: (scope: AcmeDnsProviderTableScope, id: string, options?: CreateOrEditAcmeDnsProviderDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditAcmeDnsProviderDialogState = create<State & Actions>()(set => ({
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
