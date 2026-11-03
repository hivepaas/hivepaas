import { create } from "zustand";

import type { ManagedUnmanageDialogOptions, ManagedUnmanageDialogState, ManagedUnmanageTarget } from "../types";

type State = ManagedUnmanageDialogState & ManagedUnmanageDialogOptions;

interface Actions {
    open: (target: ManagedUnmanageTarget, options?: ManagedUnmanageDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useManagedUnmanageDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        target: null,
    },
    props: {},
    open: (target, options = {}) => {
        set({
            state: {
                mode: "open",
                target,
            },
            ...options,
        });
    },
    close: () => {
        set({
            state: {
                mode: "closed",
                target: null,
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
                    target: null,
                },
                props: {},
            };
        });
    },
}));
