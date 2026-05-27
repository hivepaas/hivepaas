import { useMemo } from "react";

import type { GithubAppTableScope } from "~/settings/module-shared/components";

import type { ProvisionGithubAppDialogOptions } from "../types";

import { useProvisionGithubAppDialogState } from "./use-provision-github-app.dialog.state";

function createHook() {
    return function useProvisionGithubAppDialog(props: ProvisionGithubAppDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useProvisionGithubAppDialogState();

        return useMemo(
            () => ({
                state,
                actions: {
                    open: (scope: GithubAppTableScope, options: ProvisionGithubAppDialogOptions = {}) => {
                        actions.open(scope, { props: { ...props, ...options.props } });
                    },
                    close: actions.close,
                },
            }),
            [actions, props, state],
        );
    };
}

export const useProvisionGithubAppDialog = createHook();
