import { create } from "zustand";
import type { NotificationTargetTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditNotificationTargetDialogOptions, CreateOrEditNotificationTargetDialogState } from "../types";

type State = CreateOrEditNotificationTargetDialogState & CreateOrEditNotificationTargetDialogOptions;

interface Actions {
    open: (scope: NotificationTargetTableScope, options?: CreateOrEditNotificationTargetDialogOptions) => void;
    openEdit: (
        scope: NotificationTargetTableScope,
        id: string,
        options?: CreateOrEditNotificationTargetDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditNotificationTargetDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: {},
    open: (scope, options = {}) => {
        set({ state: { mode: "open", scope }, ...options });
    },
    openEdit: (scope, id, options = {}) => {
        set({ state: { mode: "edit", scope, id }, ...options });
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
