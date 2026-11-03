import { create } from "zustand";
import type { EmailAccountTableScope } from "~/settings/module-shared/components";

import type { UpdateEmailAccountStatusDialogOptions, UpdateEmailAccountStatusDialogState } from "../types";

type State = UpdateEmailAccountStatusDialogState & UpdateEmailAccountStatusDialogOptions;

interface Actions {
    open: (scope: EmailAccountTableScope, id: string, options?: UpdateEmailAccountStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateEmailAccountStatusDialogState = create<State & Actions>()(set => ({
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
