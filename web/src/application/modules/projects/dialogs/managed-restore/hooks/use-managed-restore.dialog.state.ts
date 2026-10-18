import { create } from "zustand";

import type {
    ManagedRestoreDialogOptions,
    ManagedRestoreDialogState,
    ManagedRestorePreset,
    ManagedRestoreSource,
} from "../types";

type State = ManagedRestoreDialogState & ManagedRestoreDialogOptions;

interface Actions {
    open: (source: ManagedRestoreSource, preset?: ManagedRestorePreset, options?: ManagedRestoreDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useManagedRestoreDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        source: null,
        preset: null,
    },
    props: {},
    open: (source, preset, options = {}) => {
        set({
            state: {
                mode: "open",
                source,
                preset: preset ?? null,
            },
            ...options,
        });
    },
    close: () => {
        set({
            state: {
                mode: "closed",
                source: null,
                preset: null,
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
                    source: null,
                    preset: null,
                },
                props: {},
            };
        });
    },
}));
