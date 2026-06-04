import { useRef } from "react";

import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectAppEnvVarsCommands } from "~/projects/data/commands";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectAppEnvVarsQueries } from "~/projects/data/queries";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigEnvVarsForm } from "../form";
import { type AppConfigEnvVarsFormSchemaOutput } from "../schemas";
import { type AppConfigEnvVarsFormRef } from "../types";

export function AppConfigEnvVariablesRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigEnvVarsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const {
        data: envVarsData,
        isLoading: isLoadingEnvVars,
        error: envVarsError,
        refetch: refetchEnvVars,
    } = ProjectAppEnvVarsQueries.useFindOne({ projectID: projectId, appID: appId }, APP_CONFIGURATION_QUERY_OPTIONS);

    const { mutate: update, isPending } = ProjectAppEnvVarsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Environment variables updated");
        },
        onError: (err: Error) => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: AppConfigEnvVarsFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");
        invariant(envVarsData, "envVarsData must be defined");

        update({
            projectID: projectId,
            appID: appId,
            ...values,
            updateVer: envVarsData.data.updateVer,
        });
    }

    if (isLoadingEnvVars) {
        return <AppLoader />;
    }

    if (envVarsError) {
        return (
            <PageError
                error={envVarsError}
                onRetry={refetchEnvVars}
            />
        );
    }

    invariant(envVarsData, "envVarsData must be defined");

    const { data: envVars } = envVarsData;

    return (
        <AppConfigEnvVarsForm
            ref={formRef}
            defaultValues={{
                buildtime: envVars.buildtime,
                runtime: envVars.runtime,
            }}
            inheritedValues={{
                buildtime: envVars.inheritedBuildtimeEnvVars,
                runtime: envVars.inheritedRuntimeEnvVars,
            }}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
        >
            <div className="flex justify-end gap-2 mt-4">
                <ProjectPermissionSubmitButton isPending={isPending} />
            </div>
        </AppConfigEnvVarsForm>
    );
}
