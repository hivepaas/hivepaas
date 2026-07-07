import { useState } from "react";

import { toast } from "sonner";
import { ProjectCloudStorageCommands } from "~/projects/data/commands";
import { ProjectCloudStorageQueries } from "~/projects/data/queries";
import { CloudStorageCommands } from "~/settings/data/commands";
import { CloudStorageQueries } from "~/settings/data/queries";
import { CreateOrEditCloudStorageForm } from "~/settings/module-shared/components/cloud-storage-form";
import type {
    CreateOrEditCloudStorageFormInput,
    CreateOrEditCloudStorageFormOutput,
} from "~/settings/module-shared/components/cloud-storage-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { ECloudStorageKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { CloudStorageTableScope } from "../cloud-storage-table";

type CloudStorageFormRouteMode = "create" | "edit";

export function CloudStorageFormRoute({ mode, scope, cloudStorageId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getCloudStorageListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (cloudStorageId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingCloudStorage, isPending: isCreatingSetting } = CloudStorageCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Cloud storage created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingCloudStorage, isPending: isUpdatingSetting } = CloudStorageCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Cloud storage updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectCloudStorage, isPending: isCreatingProject } =
        ProjectCloudStorageCommands.useCreateOne({
            onSuccess: () => {
                toast.success("Project cloud storage created successfully");
                markSaved();
            },
        });
    const { mutate: updateProjectCloudStorage, isPending: isUpdatingProject } =
        ProjectCloudStorageCommands.useUpdateOne({
            onSuccess: () => {
                toast.success("Project cloud storage updated successfully");
                markSaved();
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

    const settingDetailQuery = CloudStorageQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectCloudStorageQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const cloudStorage = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && cloudStorage?.inherited === true;

    function createPayload(values: CreateOrEditCloudStorageFormOutput) {
        return {
            availableInProjects: scope.type === "project" ? false : values.availableInProjects,
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
        const payload = createPayload(values);

        if (isEditMode && cloudStorage) {
            const updatePayload = { ...payload, updateVer: cloudStorage.updateVer };

            if (scope.type === "project") {
                updateProjectCloudStorage({
                    projectID: scope.projectId,
                    id: cloudStorage.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingCloudStorage({ id: cloudStorage.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectCloudStorage({ projectID: scope.projectId, payload });
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
        if (
            !readOnlyInherited &&
            canWrite &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        )
            return;

        navigateToList();
    }

    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditCloudStorageFormInput> | undefined = cloudStorage
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
    const shouldRenderForm = mode === "create" || initialValues;
    const title = mode === "create" ? "Create Cloud Storage" : "Edit Cloud Storage";

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditCloudStorageForm
                    isPending={isPending}
                    isTesting={isTesting}
                    testStatus={testStatus}
                    onSubmit={onSubmit}
                    onTestConnection={onTestConnection}
                    onHasChanges={setHasChanges}
                    savedVersion={saveRevision}
                    initialValues={initialValues}
                    showAvailableInProjects={scope.type === "settings"}
                    readOnlyInherited={readOnlyInherited}
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

function getCloudStorageListRoute(scope: CloudStorageTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.cloudStorages.$route(scope.projectId);
    }

    return ROUTE.settings.cloudStorages.$route;
}

interface Props {
    mode: CloudStorageFormRouteMode;
    scope: CloudStorageTableScope;
    cloudStorageId?: string;
}
