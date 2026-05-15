import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectImServiceCommands } from "~/projects/data/commands";
import { ProjectImServiceQueries } from "~/projects/data/queries";
import { ImServiceCommands } from "~/settings/data/commands";
import { ImServiceQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { EImServiceKind } from "@application/shared/enums";

import { CreateOrEditImPlatformForm } from "../form";
import { useCreateOrEditImPlatformDialogState } from "../hooks";
import type { CreateOrEditImPlatformFormOutput } from "../schemas";

export function CreateOrEditImPlatformDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useCreateOrEditImPlatformDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");

    const { mutate: createSettingImPlatform, isPending: isCreatingSetting } = ImServiceCommands.useCreateOne({
        onSuccess: () => {
            toast.success("IM platform created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: updateSettingImPlatform, isPending: isUpdatingSetting } = ImServiceCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("IM platform updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: createProjectImPlatform, isPending: isCreatingProject } = ProjectImServiceCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project IM platform created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: updateProjectImPlatform, isPending: isUpdatingProject } = ProjectImServiceCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project IM platform updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: testSendMsg, isPending: isTesting } = ImServiceCommands.useTestSendMsg({
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
    const settingDetailQuery = ImServiceQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "edit" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectImServiceQueries.useFindOneById(
        {
            projectID: state.mode === "edit" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        {
            enabled: state.mode === "edit" && state.scope.type === "project",
        },
    );
    const detailQuery =
        state.mode === "edit" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const imPlatform = detailQuery.data?.data;

    function createPayload(values: CreateOrEditImPlatformFormOutput) {
        const isSlack = values.kind === EImServiceKind.Slack;

        return {
            availableInProjects:
                state.mode !== "closed" && state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            kind: values.kind,
            slack: isSlack ? { webhook: values.webhook } : null,
            discord: isSlack ? null : { webhook: values.webhook },
        };
    }

    function onSubmit(values: CreateOrEditImPlatformFormOutput) {
        if (state.mode === "closed") {
            return;
        }

        const payload = createPayload(values);

        if (state.mode === "edit" && imPlatform) {
            const updatePayload = {
                ...payload,
                updateVer: imPlatform.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectImPlatform({
                    projectID: state.scope.projectId,
                    id: imPlatform.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingImPlatform({
                id: imPlatform.id,
                payload: updatePayload,
            });
            return;
        }

        if (state.scope.type === "project") {
            createProjectImPlatform({
                projectID: state.scope.projectId,
                payload,
            });
            return;
        }

        createSettingImPlatform({ payload });
    }

    function onTestSendMsg(values: CreateOrEditImPlatformFormOutput) {
        setTestStatus("idle");
        testSendMsg({
            payload: {
                ...createPayload(values),
                testMsg: "test message",
            },
        });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

        if (hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues = imPlatform
        ? {
              name: imPlatform.name,
              kind: imPlatform.kind,
              webhook:
                  imPlatform.kind === EImServiceKind.Slack
                      ? (imPlatform.slack?.webhook ?? "")
                      : (imPlatform.discord?.webhook ?? ""),
              availableInProjects: imPlatform.availableInProjects ?? false,
              default: imPlatform.default ?? false,
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
                    <DialogTitle>Create or update an IM platform</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode !== "closed" && !isDetailLoading && (state.mode === "open" || initialValues) && (
                    <CreateOrEditImPlatformForm
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        onSubmit={onSubmit}
                        onTestSendMsg={onTestSendMsg}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
