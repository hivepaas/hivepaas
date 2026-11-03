import { create } from "zustand";
import type { AcmeDnsProviderTableScope } from "~/settings/module-shared/components";

import type { UpdateAcmeDnsProviderStatusDialogOptions, UpdateAcmeDnsProviderStatusDialogState } from "../types";

type State = UpdateAcmeDnsProviderStatusDialogState & UpdateAcmeDnsProviderStatusDialogOptions;

interface Actions {
    open: (scope: AcmeDnsProviderTableScope, id: string, options?: UpdateAcmeDnsProviderStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateAcmeDnsProviderStatusDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (scope, id, options = {}) => {
        set({
            state: {
                mode: "open",
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
