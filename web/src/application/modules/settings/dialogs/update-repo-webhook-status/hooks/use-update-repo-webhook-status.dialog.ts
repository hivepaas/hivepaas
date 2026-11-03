import { useMemo } from "react";

import type { RepoWebhookTableScope } from "~/settings/module-shared/components";

import type { UpdateRepoWebhookStatusDialogOptions } from "../types";

import { useUpdateRepoWebhookStatusDialogState } from "./use-update-repo-webhook-status.dialog.state";

function createHook() {
    return function useUpdateRepoWebhookStatusDialog(props: UpdateRepoWebhookStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateRepoWebhookStatusDialogState();

        return useMemo(
            () => ({
                state,
                actions: {
                    open: (
                        scope: RepoWebhookTableScope,
                        id: string,
                        options: UpdateRepoWebhookStatusDialogOptions = {},
                    ) => {
                        actions.open(scope, id, { props: { ...props, ...options.props } });
                    },
                    close: actions.close,
                },
            }),
            [actions, props, state],
        );
    };
}

export const useUpdateRepoWebhookStatusDialog = createHook();
