import { useMemo } from "react";

import type { GithubAppTableScope } from "~/settings/module-shared/components";

import type { UpdateGithubAppStatusDialogOptions } from "../types";

import { useUpdateGithubAppStatusDialogState } from "./use-update-github-app-status.dialog.state";

function createHook() {
    return function useUpdateGithubAppStatusDialog(props: UpdateGithubAppStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateGithubAppStatusDialogState();

        return useMemo(
            () => ({
                state,
                actions: {
                    open: (
                        scope: GithubAppTableScope,
                        id: string,
                        options: UpdateGithubAppStatusDialogOptions = {},
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

export const useUpdateGithubAppStatusDialog = createHook();
