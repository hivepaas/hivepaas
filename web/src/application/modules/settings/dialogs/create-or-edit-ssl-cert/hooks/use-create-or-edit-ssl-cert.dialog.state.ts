import { create } from "zustand";
import type { SslCertTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditSslCertDialogOptions, CreateOrEditSslCertDialogState } from "../types";

type State = CreateOrEditSslCertDialogState & CreateOrEditSslCertDialogOptions;

interface Actions {
    open: (scope: SslCertTableScope, options?: CreateOrEditSslCertDialogOptions) => void;
    openEdit: (scope: SslCertTableScope, id: string, options?: CreateOrEditSslCertDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditSslCertDialogState = create<State & Actions>()(set => ({
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
