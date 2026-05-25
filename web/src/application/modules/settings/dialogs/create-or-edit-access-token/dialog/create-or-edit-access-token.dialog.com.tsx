import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectAccessTokenCommands } from "~/projects/data/commands";
import { ProjectAccessTokenQueries } from "~/projects/data/queries";
import { AccessTokenCommands } from "~/settings/data/commands";
import { AccessTokenQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { EAccessTokenKind } from "@application/shared/enums";

import { CreateOrEditAccessTokenForm } from "../form";
import { useCreateOrEditAccessTokenDialogState } from "../hooks";
import type { CreateOrEditAccessTokenFormOutput } from "../schemas";

export function CreateOrEditAccessTokenDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useCreateOrEditAccessTokenDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");

    const { mutate: createSettingAccessToken, isPending: isCreatingSetting } = AccessTokenCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Access token created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateSettingAccessToken, isPending: isUpdatingSetting } = AccessTokenCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Access token updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: createProjectAccessToken, isPending: isCreatingProject } = ProjectAccessTokenCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project access token created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateProjectAccessToken, isPending: isUpdatingProject } = ProjectAccessTokenCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project access token updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: testConnection, isPending: isTesting } = AccessTokenCommands.useTestConn({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            setTestStatus("idle");
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "edit" ? state.id : "";
    const settingDetailQuery = AccessTokenQueries.useFindOneById(
        { id: detailId },
        { enabled: state.mode === "edit" && state.scope.type === "settings" },
    );
    const projectDetailQuery = ProjectAccessTokenQueries.useFindOneById(
        {
            projectID: state.mode === "edit" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        { enabled: state.mode === "edit" && state.scope.type === "project" },
    );
    const detailQuery =
        state.mode === "edit" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const accessToken = detailQuery.data?.data;

    function createPayload(values: CreateOrEditAccessTokenFormOutput) {
        return {
            availableInProjects:
                state.mode !== "closed" && state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            expireAt: values.expireAt,
            kind: values.kind,
            name: values.name,
            user: values.user,
            token: values.token,
            baseURL: values.baseURL,
        };
    }

    function onSubmit(values: CreateOrEditAccessTokenFormOutput) {
        if (state.mode === "closed") return;
        const payload = createPayload(values);

        if (state.mode === "edit" && accessToken) {
            const updatePayload = { ...payload, updateVer: accessToken.updateVer };
            if (state.scope.type === "project") {
                updateProjectAccessToken({
                    projectID: state.scope.projectId,
                    id: accessToken.id,
                    payload: updatePayload,
                });
                return;
            }
            updateSettingAccessToken({ id: accessToken.id, payload: updatePayload });
            return;
        }

        if (state.scope.type === "project") {
            createProjectAccessToken({ projectID: state.scope.projectId, payload });
            return;
        }

        createSettingAccessToken({ payload });
    }

    function onTestConnection(values: CreateOrEditAccessTokenFormOutput) {
        setTestStatus("idle");
        testConnection({
            payload: {
                expireAt: values.expireAt,
                kind: values.kind,
                name: values.name,
                user: values.user,
                token: values.token,
                baseURL: values.baseURL,
            },
        });
    }

    function handleClose() {
        if (isPending) return;
        if (
            !readOnlyInherited &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        )
            return;
        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const resolvedDialogOptions = dialogOptions ?? {};
    const readOnlyInherited = resolvedDialogOptions.readOnlyInherited === true;
    const dialogTitle = readOnlyInherited
        ? (resolvedDialogOptions.entityTitle ?? "Access Token")
        : "Create or update an access token";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues = accessToken
        ? {
              name: accessToken.name,
              kind: accessToken.kind ?? EAccessTokenKind.Github,
              user: accessToken.user,
              token: accessToken.token,
              baseURL: accessToken.baseURL,
              expireAt: accessToken.expireAt ?? null,
              availableInProjects: accessToken.availableInProjects ?? false,
              default: accessToken.default ?? false,
          }
        : undefined;
    const isDetailLoading = state.mode === "edit" && detailQuery.isFetching;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[760px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>{dialogTitle}</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode !== "closed" && !isDetailLoading && (state.mode === "open" || initialValues) && (
                    <CreateOrEditAccessTokenForm
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        onSubmit={onSubmit}
                        onTestConnection={onTestConnection}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                        readOnlyInherited={readOnlyInherited}
                        onClose={handleClose}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
