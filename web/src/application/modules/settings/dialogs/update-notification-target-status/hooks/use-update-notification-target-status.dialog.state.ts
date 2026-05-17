import { create } from "zustand";
import type { NotificationTargetTableScope } from "~/settings/module-shared/components";

import type { UpdateNotificationTargetStatusDialogOptions, UpdateNotificationTargetStatusDialogState } from "../types";

type State = UpdateNotificationTargetStatusDialogState & UpdateNotificationTargetStatusDialogOptions;

interface Actions {
    open: (
        scope: NotificationTargetTableScope,
        id: string,
        options?: UpdateNotificationTargetStatusDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateNotificationTargetStatusDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: {},
    open: (scope, id, options = {}) => {
        set({ state: { mode: "open", scope, id }, ...options });
    },
    close: () => {
        set({ state: { mode: "closed" } });
    },
    clear: () => {
        set({ props: {} });
    },
    destroy: () => {
        set(state => {
            if (state.state.mode === "closed") return state;
            return { state: { mode: "closed" }, props: {} };
        });
    },
}));
