import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectSslProviderCommands } from "~/projects/data/commands";
import { ProjectSslProviderQueries } from "~/projects/data/queries";
import type { SslProvider_CreateOne_Payload } from "~/settings/api/services/ssl-provider-services";
import { SslProviderCommands } from "~/settings/data/commands";
import { SslProviderQueries } from "~/settings/data/queries";
import type { SettingSslProvider } from "~/settings/domain";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ESslProviderKind } from "@application/shared/enums";

import { CreateOrEditSslProviderForm } from "../form";
import { useCreateOrEditSslProviderDialogState } from "../hooks";
import {
    type CreateOrEditSslProviderFormInput,
    type CreateOrEditSslProviderFormOutput,
    SSL_PROVIDER_UNSPECIFIED_KEY_TYPE,
} from "../schemas";

function getEabValues(sslProvider?: SettingSslProvider) {
    if (!sslProvider) {
        return {
            eabKid: "",
            eabHmacKey: "",
        };
    }

    if (sslProvider.kind === ESslProviderKind.ZeroSSL) {
        return {
            eabKid: sslProvider.zeroSSL?.eabKid ?? "",
            eabHmacKey: sslProvider.zeroSSL?.eabHmacKey ?? "",
        };
    }

    if (sslProvider.kind === ESslProviderKind.GoogleTrustServices) {
        return {
            eabKid: sslProvider.googleTS?.eabKid ?? "",
            eabHmacKey: sslProvider.googleTS?.eabHmacKey ?? "",
        };
    }

    return {
        eabKid: "",
        eabHmacKey: "",
    };
}

function createPayload(
    values: CreateOrEditSslProviderFormOutput,
    isProjectScope: boolean,
): SslProvider_CreateOne_Payload {
    const defaultKeyType =
        values.defaultKeyType === SSL_PROVIDER_UNSPECIFIED_KEY_TYPE ? undefined : values.defaultKeyType;
    const isLetsEncrypt = values.kind === ESslProviderKind.LetsEncrypt;
    const isZeroSSL = values.kind === ESslProviderKind.ZeroSSL;
    const isGoogleTS = values.kind === ESslProviderKind.GoogleTrustServices;
    const eabPayload = {
        eabKid: values.eabKid,
        eabHmacKey: values.eabHmacKey,
    };

    return {
        availableInProjects: isProjectScope ? false : values.availableInProjects,
        default: values.default,
        name: values.name,
        kind: values.kind,
        email: values.email,
        ...(defaultKeyType ? { defaultKeyType } : {}),
        letsEncrypt: isLetsEncrypt ? {} : null,
        zeroSSL: isZeroSSL ? eabPayload : null,
        googleTS: isGoogleTS ? eabPayload : null,
    };
}

export function CreateOrEditSslProviderDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useCreateOrEditSslProviderDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const permissionScope = state.mode === "closed" ? ({ type: "settings" } as const) : state.scope;
    const { canWrite } = useSettingsScopePermissions(permissionScope);

    const { mutate: createSettingSslProvider, isPending: isCreatingSetting } = SslProviderCommands.useCreateOne({
        onSuccess: () => {
            toast.success("SSL provider created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: updateSettingSslProvider, isPending: isUpdatingSetting } = SslProviderCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("SSL provider updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: createProjectSslProvider, isPending: isCreatingProject } = ProjectSslProviderCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project SSL provider created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });
    const { mutate: updateProjectSslProvider, isPending: isUpdatingProject } = ProjectSslProviderCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project SSL provider updated successfully");
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

    const detailId = state.mode === "edit" ? state.id : "";
    const settingDetailQuery = SslProviderQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "edit" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectSslProviderQueries.useFindOneById(
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
    const sslProvider = detailQuery.data?.data;

    function onSubmit(values: CreateOrEditSslProviderFormOutput) {
        if (state.mode === "closed") {
            return;
        }

        const payload = createPayload(values, state.scope.type === "project");

        if (state.mode === "edit" && sslProvider) {
            const updatePayload = {
                ...payload,
                updateVer: sslProvider.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectSslProvider({
                    projectID: state.scope.projectId,
                    id: sslProvider.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingSslProvider({
                id: sslProvider.id,
                payload: updatePayload,
            });
            return;
        }

        if (state.scope.type === "project") {
            createProjectSslProvider({
                projectID: state.scope.projectId,
                payload,
            });
            return;
        }

        createSettingSslProvider({ payload });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

        if (
            !readOnlyInherited &&
            canWrite &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        ) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const resolvedDialogOptions = dialogOptions ?? {};
    const readOnlyInherited = resolvedDialogOptions.readOnlyInherited === true;
    const dialogTitle = readOnlyInherited
        ? (resolvedDialogOptions.entityTitle ?? "SSL Provider")
        : "Create or update an SSL provider";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues: Partial<CreateOrEditSslProviderFormInput> | undefined =
        state.mode === "edit" && sslProvider
            ? {
                  name: sslProvider.name,
                  kind: sslProvider.kind,
                  email: sslProvider.email,
                  defaultKeyType: sslProvider.defaultKeyType || SSL_PROVIDER_UNSPECIFIED_KEY_TYPE,
                  ...getEabValues(sslProvider),
                  availableInProjects: sslProvider.availableInProjects ?? false,
                  default: sslProvider.default ?? false,
              }
            : {
                  kind: ESslProviderKind.LetsEncrypt,
                  defaultKeyType: SSL_PROVIDER_UNSPECIFIED_KEY_TYPE,
                  availableInProjects: false,
                  default: false,
              };
    const isDetailLoading = state.mode === "edit" && detailQuery.isFetching;
    const canRenderForm = state.mode === "open" || (state.mode === "edit" && !!sslProvider);

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="min-w-[390px] w-[760px]">
                <DialogHeader>
                    <DialogTitle>{dialogTitle}</DialogTitle>
                </DialogHeader>
                {isDetailLoading && (
                    <DialogBody>
                        <AppLoader />
                    </DialogBody>
                )}
                {state.mode !== "closed" && !isDetailLoading && canRenderForm && (
                    <CreateOrEditSslProviderForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                        isEdit={state.mode === "edit"}
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
