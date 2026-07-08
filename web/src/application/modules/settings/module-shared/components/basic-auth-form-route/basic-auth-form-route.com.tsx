import { useState } from "react";

import { toast } from "sonner";
import { ProjectBasicAuthCommands } from "~/projects/data/commands";
import { ProjectBasicAuthQueries } from "~/projects/data/queries";
import { BasicAuthCommands } from "~/settings/data/commands";
import { BasicAuthQueries } from "~/settings/data/queries";
import { CreateOrEditBasicAuthForm } from "~/settings/module-shared/components/basic-auth-form";
import type {
    CreateOrEditBasicAuthFormInput,
    CreateOrEditBasicAuthFormOutput,
} from "~/settings/module-shared/components/basic-auth-form";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";

import { AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { BasicAuthTableScope } from "../basic-auth-table";

type BasicAuthFormRouteMode = "create" | "edit";

export function BasicAuthFormRoute({ mode, scope, basicAuthId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useSettingsScopePermissions(scope);
    const { navigate } = useAppNavigate();

    const listRoute = getBasicAuthListRoute(scope);
    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (basicAuthId ?? "") : "";

    function navigateToList() {
        navigate.modules(listRoute, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
    }

    const { mutate: createSettingBasicAuth, isPending: isCreatingSetting } = BasicAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Basic auth created successfully");
            markSaved();
        },
    });
    const { mutate: updateSettingBasicAuth, isPending: isUpdatingSetting } = BasicAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Basic auth updated successfully");
            markSaved();
        },
    });
    const { mutate: createProjectBasicAuth, isPending: isCreatingProject } = ProjectBasicAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project basic auth created successfully");
            markSaved();
        },
    });
    const { mutate: updateProjectBasicAuth, isPending: isUpdatingProject } = ProjectBasicAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project basic auth updated successfully");
            markSaved();
        },
    });

    const settingDetailQuery = BasicAuthQueries.useFindOneById(
        { id: detailId },
        { enabled: isEditMode && scope.type === "settings" },
    );
    const projectDetailQuery = ProjectBasicAuthQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            id: detailId,
        },
        { enabled: isEditMode && scope.type === "project" },
    );
    const detailQuery = scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const basicAuth = detailQuery.data?.data;
    const readOnlyInherited = scope.type === "project" && basicAuth?.inherited === true;

    function createPayload(values: CreateOrEditBasicAuthFormOutput) {
        return {
            availableInProjects: scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            username: values.username,
            password: values.password,
        };
    }

    function onSubmit(values: CreateOrEditBasicAuthFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && basicAuth) {
            const updatePayload = { ...payload, updateVer: basicAuth.updateVer };

            if (scope.type === "project") {
                updateProjectBasicAuth({
                    projectID: scope.projectId,
                    id: basicAuth.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingBasicAuth({ id: basicAuth.id, payload: updatePayload });
            return;
        }

        if (scope.type === "project") {
            createProjectBasicAuth({ projectID: scope.projectId, payload });
            return;
        }

        createSettingBasicAuth({ payload });
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
    const initialValues: Partial<CreateOrEditBasicAuthFormInput> | undefined = basicAuth
        ? {
              name: basicAuth.name,
              username: basicAuth.username,
              password: basicAuth.password,
              availableInProjects: basicAuth.availableInProjects ?? false,
              default: basicAuth.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || initialValues;
    const title = mode === "create" ? "Create Basic Auth" : "Edit Basic Auth";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditBasicAuthForm
                    isPending={isPending}
                    onSubmit={onSubmit}
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

function getBasicAuthListRoute(scope: BasicAuthTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.basicAuth.$route(scope.projectId);
    }

    return ROUTE.settings.basicAuth.$route;
}

interface Props {
    mode: BasicAuthFormRouteMode;
    scope: BasicAuthTableScope;
    basicAuthId?: string;
}
