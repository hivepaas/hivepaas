import { useMemo } from "react";

import type { RepoWebhookTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditRepoWebhookDialogOptions } from "../types";

import { useCreateOrEditRepoWebhookDialogState } from "./use-create-or-edit-repo-webhook.dialog.state";

function createHook() {
    return function useCreateOrEditRepoWebhookDialog(props: CreateOrEditRepoWebhookDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditRepoWebhookDialogState();

        return useMemo(
            () => ({
                state,
                actions: {
                    open: (scope: RepoWebhookTableScope, options: CreateOrEditRepoWebhookDialogOptions = {}) => {
                        actions.open(scope, { props: { ...props, ...options.props } });
                    },
                    openEdit: (
                        scope: RepoWebhookTableScope,
                        id: string,
                        options: CreateOrEditRepoWebhookDialogOptions = {},
                    ) => {
                        actions.openEdit(scope, id, { props: { ...props, ...options.props } });
                    },
                    close: actions.close,
                },
            }),
            [actions, props, state],
        );
    };
}

export const useCreateOrEditRepoWebhookDialog = createHook();
