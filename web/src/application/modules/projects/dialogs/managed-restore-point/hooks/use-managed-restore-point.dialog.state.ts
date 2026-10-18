import { create } from "zustand";
import type { AppManagedRestorePoint } from "~/projects/domain";

import type {
    ManagedRestorePointDialogOptions,
    ManagedRestorePointDialogState,
    ManagedRestorePointTarget,
} from "../types";

type State = ManagedRestorePointDialogState & ManagedRestorePointDialogOptions;

interface Actions {
    openCreate: (target: ManagedRestorePointTarget, options?: ManagedRestorePointDialogOptions) => void;
    openEdit: (
        target: ManagedRestorePointTarget,
        point: AppManagedRestorePoint,
        options?: ManagedRestorePointDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useManagedRestorePointDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        target: null,
        point: null,
    },
    props: {},
    openCreate: (target, options = {}) => {
        set({
            state: {
                mode: "create",
                target,
                point: null,
            },
            ...options,
        });
    },
    openEdit: (target, point, options = {}) => {
        set({
            state: {
                mode: "edit",
                target,
                point,
            },
            ...options,
        });
    },
    close: () => {
        set({
            state: {
                mode: "closed",
                target: null,
                point: null,
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
                    point: null,
                },
                props: {},
            };
        });
    },
}));
