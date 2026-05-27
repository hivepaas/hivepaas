import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { ProjectGithubAppCommands } from "~/projects/data/commands";
import { GithubAppCommands } from "~/settings/data/commands";
import { EGithubAppOwnerType } from "~/settings/module-shared/enums";

import { ProvisionGithubAppForm } from "../form";
import { useProvisionGithubAppDialogState } from "../hooks";
import type { ProvisionGithubAppFormOutput } from "../schemas";

export function ProvisionGithubAppDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useProvisionGithubAppDialogState();
    const [loginChecked, setLoginChecked] = useState(false);

    const { mutate: beginSettingsManifestFlow, isPending: isBeginningSettings } =
        GithubAppCommands.useBeginManifestFlow({
            onSuccess: response => {
                dialogOptions?.onSuccess?.();
                window.location.assign(response.data.redirectURL);
            },
        });
    const { mutate: beginProjectManifestFlow, isPending: isBeginningProject } =
        ProjectGithubAppCommands.useBeginManifestFlow({
            onSuccess: response => {
                dialogOptions?.onSuccess?.();
                window.location.assign(response.data.redirectURL);
            },
        });

    useEffect(() => {
        if (state.mode === "closed") {
            setLoginChecked(false);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    function onLoginCheck() {
        window.open("https://github.com/login", "_blank", "noopener,noreferrer");
        setLoginChecked(true);
    }

    function onSubmit(values: ProvisionGithubAppFormOutput) {
        if (state.mode !== "open") {
            return;
        }

        const payload = {
            name: values.name,
            org: values.ownerType === EGithubAppOwnerType.Organization ? values.org : "",
            ssoEnabled: values.ssoEnabled,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            beginProjectManifestFlow({
                projectID: state.scope.projectId,
                payload,
            });
            return;
        }

        beginSettingsManifestFlow({ payload });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const isPending = isBeginningSettings || isBeginningProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[840px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Provision Github app</DialogTitle>
                </DialogHeader>
                {state.mode === "open" && (
                    <ProvisionGithubAppForm
                        isPending={isPending}
                        loginChecked={loginChecked}
                        onLoginCheck={onLoginCheck}
                        onSubmit={onSubmit}
                        showAvailableInProjects={showAvailableInProjects}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
