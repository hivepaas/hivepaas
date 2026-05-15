import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectSslCertCommands } from "~/projects/data/commands";
import { ProjectDomainSettingsQueries, ProjectSslCertQueries } from "~/projects/data/queries";
import type { SslCert_Notification_Payload } from "~/settings/api/services/ssl-cert-services";
import { DomainSettingsQueries, SslCertCommands, SslCertQueries } from "~/settings/data";

import { AppLoader } from "@application/shared/components";
import { ESslCertType, ESslKeyType } from "@application/shared/enums";

import { CreateOrEditSslCertForm } from "../form";
import { useCreateOrEditSslCertDialogState } from "../hooks";
import type { CreateOrEditSslCertFormInput, CreateOrEditSslCertFormOutput } from "../schemas";

function addDays(date: Date, days: number): Date {
    const result = new Date(date);
    result.setDate(result.getDate() + days);
    return result;
}

function normalizeCertType(certType?: string): typeof ESslCertType.LetsEncrypt | typeof ESslCertType.Custom {
    return certType === ESslCertType.LetsEncrypt ? ESslCertType.LetsEncrypt : ESslCertType.Custom;
}

function buildNotificationPayload(
    notification: CreateOrEditSslCertFormOutput["notification"],
): SslCert_Notification_Payload {
    return {
        success: { id: notification.successUseDefault ? "" : (notification.success?.id ?? "") },
        successUseDefault: notification.successUseDefault,
        failure: { id: notification.failureUseDefault ? "" : (notification.failure?.id ?? "") },
        failureUseDefault: notification.failureUseDefault,
    };
}

export function CreateOrEditSslCertDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useCreateOrEditSslCertDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const open = state.mode !== "closed";
    const projectId = state.mode !== "closed" && state.scope.type === "project" ? state.scope.projectId : "";

    const settingsDomainSettingsQuery = DomainSettingsQueries.useFindOne(
        {},
        {
            enabled: open && state.mode === "open" && state.scope.type === "settings",
        },
    );

    const projectDomainSettingsQuery = ProjectDomainSettingsQueries.useFindOne(
        {
            projectID: projectId,
        },
        {
            enabled: open && state.mode === "open" && state.scope.type === "project" && !!projectId,
        },
    );

    const detailId = state.mode === "edit" ? state.id : "";
    const settingDetailQuery = SslCertQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "edit" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectSslCertQueries.useFindOneById(
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
    const sslCert = detailQuery.data?.data;

    const { mutate: createSettingSslCert, isPending: isCreatingSetting } = SslCertCommands.useCreateOne({
        onSuccess: () => {
            toast.success("SSL certificate created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: updateSettingSslCert, isPending: isUpdatingSetting } = SslCertCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("SSL certificate updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: createProjectSslCert, isPending: isCreatingProject } = ProjectSslCertCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project SSL certificate created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: updateProjectSslCert, isPending: isUpdatingProject } = ProjectSslCertCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project SSL certificate updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    function createPayload(values: CreateOrEditSslCertFormOutput) {
        const now = new Date();
        const isLetsEncrypt = values.certType === ESslCertType.LetsEncrypt;
        const fallbackExpireAt = addDays(now, isLetsEncrypt ? 90 : 365);
        const expireAt = values.expireAt ?? fallbackExpireAt;
        const notifyFrom = values.notifyFrom ?? addDays(expireAt, -30);

        return {
            availableInProjects:
                state.mode !== "closed" && state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            certType: values.certType,
            domain: values.domain,
            certificate: isLetsEncrypt ? "" : values.certificate,
            privateKey: isLetsEncrypt ? "" : values.privateKey,
            keyType: values.keyType,
            validPeriod: isLetsEncrypt ? "90d" : "365d",
            email: values.email,
            autoRenew: isLetsEncrypt ? values.autoRenew : false,
            expireAt,
            notifyFrom,
            notification: buildNotificationPayload(values.notification),
        };
    }

    function onSubmit(values: CreateOrEditSslCertFormOutput) {
        if (state.mode === "closed") {
            return;
        }

        const payload = createPayload(values);

        if (state.mode === "edit" && sslCert) {
            const updatePayload = {
                ...payload,
                updateVer: sslCert.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectSslCert({
                    projectID: state.scope.projectId,
                    id: sslCert.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingSslCert({
                id: sslCert.id,
                payload: updatePayload,
            });
            return;
        }

        if (state.scope.type === "project") {
            createProjectSslCert({
                projectID: state.scope.projectId,
                payload,
            });
            return;
        }

        createSettingSslCert({ payload });
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

    const domainSettings =
        state.mode !== "closed" && state.scope.type === "project"
            ? projectDomainSettingsQuery.data?.data.certSettings
            : settingsDomainSettingsQuery.data?.data.certSettings;
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues: Partial<CreateOrEditSslCertFormInput> | undefined =
        state.mode === "edit" && sslCert
            ? {
                  domain: sslCert.domain,
                  certType: normalizeCertType(sslCert.certType),
                  email: sslCert.email,
                  keyType: sslCert.keyType,
                  autoRenew: sslCert.autoRenew,
                  certificate: sslCert.certificate,
                  privateKey: sslCert.privateKey,
                  expireAt: sslCert.expireAt ?? null,
                  notifyFrom: sslCert.notifyFrom ?? null,
                  availableInProjects: sslCert.availableInProjects ?? false,
                  default: sslCert.default ?? false,
                  notification: {
                      successUseDefault: sslCert.notification?.successUseDefault ?? true,
                      success: sslCert.notification?.success ?? undefined,
                      failureUseDefault: sslCert.notification?.failureUseDefault ?? true,
                      failure: sslCert.notification?.failure ?? undefined,
                  },
              }
            : {
                  certType: ESslCertType.LetsEncrypt,
                  email: domainSettings?.email ?? "",
                  keyType: domainSettings?.keyType ?? ESslKeyType.ECP256,
                  autoRenew: domainSettings?.autoRenew ?? true,
                  availableInProjects: false,
                  default: false,
                  notification: {
                      successUseDefault: true,
                      failureUseDefault: true,
                  },
              };
    const isDetailLoading = state.mode === "edit" && detailQuery.isFetching;
    const canRenderForm = state.mode === "open" || (state.mode === "edit" && !!sslCert);

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[760px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create or update an SSL certificate</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode !== "closed" && !isDetailLoading && canRenderForm && (
                    <CreateOrEditSslCertForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        scope={state.scope}
                        showAvailableInProjects={showAvailableInProjects}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
