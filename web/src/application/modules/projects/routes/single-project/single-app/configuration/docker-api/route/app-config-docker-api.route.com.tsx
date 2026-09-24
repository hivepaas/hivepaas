import { useRef } from "react";

import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppDockerApiSettingsCommands, AppDockerApiSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigDockerApiForm, mapDockerApiFormValuesToPayload } from "../form";
import { type AppConfigDockerApiFormSchemaOutput } from "../schemas";
import { type AppConfigDockerApiFormRef } from "../types";

export function AppConfigDockerApiRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const formRef = useRef<AppConfigDockerApiFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { canWrite: canWriteCluster } = useConditionalModule({ id: MODULE_IDS.Cluster });

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppDockerApiSettingsQueries.useFindOne(
        {
            projectID: projectId,
            env,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const { mutate: update, isPending } = AppDockerApiSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Docker API settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update Docker API settings");
            }
        },
    });

    function handleSubmit(values: AppConfigDockerApiFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        invariant(projectId, "projectId must be defined");
        invariant(env, "env must be defined");
        invariant(appId, "appId must be defined");

        update({
            projectID: projectId,
            env,
            appID: appId,
            payload: mapDockerApiFormValuesToPayload(values, data?.data.updateVer ?? 0),
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <AppConfigDockerApiForm
                ref={formRef}
                defaultValues={data?.data}
                canWriteCluster={canWriteCluster}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
            >
                <FormActionBar>
                    <ProjectPermissionSubmitButton isPending={isPending} />
                </FormActionBar>
            </AppConfigDockerApiForm>
        </div>
    );
}
