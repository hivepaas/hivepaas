import { useMemo } from "react";

import type { GithubAppTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditGithubAppDialogOptions } from "../types";

import { useCreateOrEditGithubAppDialogState } from "./use-create-or-edit-github-app.dialog.state";

function createHook() {
    return function useCreateOrEditGithubAppDialog(props: CreateOrEditGithubAppDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditGithubAppDialogState();

        return useMemo(
            () => ({
                state,
                actions: {
                    open: (scope: GithubAppTableScope, options: CreateOrEditGithubAppDialogOptions = {}) => {
                        actions.open(scope, { props: { ...props, ...options.props } });
                    },
                    openEdit: (
                        scope: GithubAppTableScope,
                        id: string,
                        options: CreateOrEditGithubAppDialogOptions = {},
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

export const useCreateOrEditGithubAppDialog = createHook();
