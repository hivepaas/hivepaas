import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectCloudStorageCommands } from "~/projects/data/commands";
import { ProjectCloudStorageQueries } from "~/projects/data/queries";
import { CloudStorageCommands } from "~/settings/data/commands";
import { CloudStorageQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { ECloudStorageKind } from "@application/shared/enums";

import { CreateOrEditCloudStorageForm } from "../form";
import { useCreateOrEditCloudStorageDialogState } from "../hooks";
import type { CreateOrEditCloudStorageFormOutput } from "../schemas";

export function CreateOrEditCloudStorageDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useCreateOrEditCloudStorageDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");

    const { mutate: createSettingCloudStorage, isPending: isCreatingSetting } = CloudStorageCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Cloud storage created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateSettingCloudStorage, isPending: isUpdatingSetting } = CloudStorageCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Cloud storage updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: createProjectCloudStorage, isPending: isCreatingProject } =
        ProjectCloudStorageCommands.useCreateOne({
            onSuccess: () => {
                toast.success("Project cloud storage created successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
        });
    const { mutate: updateProjectCloudStorage, isPending: isUpdatingProject } =
        ProjectCloudStorageCommands.useUpdateOne({
            onSuccess: () => {
                toast.success("Project cloud storage updated successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
        });
    const { mutate: testConnection, isPending: isTesting } = CloudStorageCommands.useTestConn({
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
    const settingDetailQuery = CloudStorageQueries.useFindOneById(
        { id: detailId },
        { enabled: state.mode === "edit" && state.scope.type === "settings" },
    );
    const projectDetailQuery = ProjectCloudStorageQueries.useFindOneById(
        {
            projectID: state.mode === "edit" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        { enabled: state.mode === "edit" && state.scope.type === "project" },
    );
    const detailQuery =
        state.mode === "edit" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const cloudStorage = detailQuery.data?.data;

    function createPayload(values: CreateOrEditCloudStorageFormOutput) {
        return {
            availableInProjects:
                state.mode !== "closed" && state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            kind: values.kind,
            name: values.name,
            s3: {
                accessKeyId: values.accessKeyId,
                secretKey: values.secretKey,
                region: values.region,
                bucket: values.bucket,
                endpoint: values.endpoint,
            },
        };
    }

    function onSubmit(values: CreateOrEditCloudStorageFormOutput) {
        if (state.mode === "closed") return;
        const payload = createPayload(values);

        if (state.mode === "edit" && cloudStorage) {
            const updatePayload = { ...payload, updateVer: cloudStorage.updateVer };
            if (state.scope.type === "project") {
                updateProjectCloudStorage({
                    projectID: state.scope.projectId,
                    id: cloudStorage.id,
                    payload: updatePayload,
                });
                return;
            }
            updateSettingCloudStorage({ id: cloudStorage.id, payload: updatePayload });
            return;
        }

        if (state.scope.type === "project") {
            createProjectCloudStorage({ projectID: state.scope.projectId, payload });
            return;
        }

        createSettingCloudStorage({ payload });
    }

    function onTestConnection(values: CreateOrEditCloudStorageFormOutput) {
        setTestStatus("idle");
        testConnection({
            payload: {
                kind: values.kind,
                name: values.name,
                s3: {
                    accessKeyId: values.accessKeyId,
                    secretKey: values.secretKey,
                    region: values.region,
                    bucket: values.bucket,
                    endpoint: values.endpoint,
                },
            },
        });
    }

    function handleClose() {
        if (isPending) return;
        if (hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) return;
        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues = cloudStorage
        ? {
              name: cloudStorage.name,
              kind: cloudStorage.kind === ECloudStorageKind.AWSS3 ? ECloudStorageKind.AWSS3 : ECloudStorageKind.AWSS3,
              accessKeyId: cloudStorage.s3.accessKeyId,
              secretKey: cloudStorage.s3.secretKey,
              region: cloudStorage.s3.region,
              bucket: cloudStorage.s3.bucket,
              endpoint: cloudStorage.s3.endpoint,
              availableInProjects: cloudStorage.availableInProjects ?? false,
              default: cloudStorage.default ?? false,
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
                    <DialogTitle>Create or update a Cloud Storage</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode !== "closed" && !isDetailLoading && (state.mode === "open" || initialValues) && (
                    <CreateOrEditCloudStorageForm
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        onSubmit={onSubmit}
                        onTestConnection={onTestConnection}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
