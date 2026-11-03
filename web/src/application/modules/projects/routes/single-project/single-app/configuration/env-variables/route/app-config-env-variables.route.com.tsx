import { useMemo, useRef } from "react";

import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectAppEnvVarsCommands } from "~/projects/data/commands";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectAppEnvVarsQueries } from "~/projects/data/queries";
import type { ProjectRuntimeEnvVar } from "~/projects/domain";
import { ManagedLockNotice, ProjectPermissionSubmitButton } from "~/projects/module-shared/components";
import { useAppManagedLocks } from "~/projects/module-shared/hooks";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigEnvVarsForm } from "../form";
import { type AppConfigEnvVarsFormSchemaOutput } from "../schemas";
import { type AppConfigEnvVarsFormRef } from "../types";

export function AppConfigEnvVariablesRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const formRef = useRef<AppConfigEnvVarsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const {
        data: envVarsData,
        isLoading: isLoadingEnvVars,
        error: envVarsError,
        refetch: refetchEnvVars,
    } = ProjectAppEnvVarsQueries.useFindOne(
        { projectID: projectId, env, appID: appId },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const managed = useAppManagedLocks({ projectID: projectId, env, appID: appId });
    const lockedKeys = managed?.locks.envKeys;
    const defaultValues = useMemo(() => {
        if (!envVarsData) {
            return undefined;
        }
        // A key the managed database's engine owns is shown, and kept, as it is.
        const lock = (vars: ProjectRuntimeEnvVar[]) =>
            vars.map(item => (lockedKeys?.includes(item.key) ? { ...item, isReadOnly: true } : item));

        return {
            buildtime: envVarsData.data.buildtime,
            runtime: lock(envVarsData.data.runtime),
            shared: envVarsData.data.shared,
        };
    }, [envVarsData, lockedKeys]);

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
        invariant(env, "env must be defined");
        invariant(appId, "appId must be defined");
        invariant(envVarsData, "envVarsData must be defined");

        update({
            projectID: projectId,
            env,
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
    invariant(defaultValues, "defaultValues must be defined");

    const { data: envVars } = envVarsData;

    return (
        <AppConfigEnvVarsForm
            ref={formRef}
            defaultValues={defaultValues}
            inheritedValues={{
                buildtime: envVars.inheritedBuildtimeEnvVars,
                runtime: envVars.inheritedRuntimeEnvVars,
            }}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
        >
            {managed && managed.locks.envKeys.length > 0 && (
                <ManagedLockNotice
                    projectId={projectId}
                    env={env}
                    appId={appId}
                >
                    {managed.locks.envKeys.join(", ")} keep their values.
                </ManagedLockNotice>
            )}
            <FormActionBar contentClassName="gap-2">
                <ProjectPermissionSubmitButton isPending={isPending} />
            </FormActionBar>
        </AppConfigEnvVarsForm>
    );
}
