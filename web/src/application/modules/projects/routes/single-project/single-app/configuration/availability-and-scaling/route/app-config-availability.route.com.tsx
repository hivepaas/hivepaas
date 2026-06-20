import { useRef } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppServiceSettingsCommands, AppServiceSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLink, AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { EServiceMode } from "@application/modules/projects/module-shared/enums";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigAvailabilityForm } from "../form";
import { type AppConfigAvailabilitySchemaOutput } from "../schemas";
import { type AppConfigAvailabilityFormRef } from "../types";

export function AppConfigAvailabilityRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigAvailabilityFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppServiceSettingsQueries.useFindOne(
        {
            projectID: projectId,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const { mutate: update, isPending } = AppServiceSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Availability and scaling settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update availability and scaling settings");
            }
        },
    });

    function handleSubmit(values: AppConfigAvailabilitySchemaOutput) {
        if (!canWrite) {
            return;
        }

        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");

        update({
            projectID: projectId,
            appID: appId,
            payload: {
                updateVer: data?.data.updateVer ?? 0,
                modeSpec: {
                    mode: values.mode,
                    serviceReplicas: values.mode === EServiceMode.Replicated ? values.serviceReplicas : null,
                    jobMaxConcurrent: values.mode === EServiceMode.ReplicatedJob ? values.jobMaxConcurrent : null,
                    jobTotalCompletions: values.mode === EServiceMode.ReplicatedJob ? values.jobTotalCompletions : null,
                },
                placement: {
                    constraints: values.constraints,
                    preferences: values.preferences,
                },
            },
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <div className={cn(dashedBorderBox, "text-sm p-2 leading-6 text-muted-foreground")}>
                For configuration details, see{" "}
                <a
                    href="https://docs.docker.com/reference/api/engine/version/v1.52/#tag/Service/operation/ServiceUpdate"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-500 underline underline-offset-2"
                >
                    docs
                </a>
            </div>

            <div className={cn(dashedBorderBox, "text-center text-sm p-2 leading-6")}>
                <span className="text-orange-500">Note:</span> If you change the configuration here, please check the
                application&rsquo;s scheduling results in{" "}
                <AppLink.Basic
                    to={ROUTE.projects.single.apps.single.instances.$route(projectId, appId)}
                    className="text-primary underline-offset-4 hover:underline"
                >
                    Instances
                </AppLink.Basic>
                .
            </div>

            <AppConfigAvailabilityForm
                ref={formRef}
                defaultValues={data?.data}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
            >
                <div className="flex justify-end mt-4">
                    <ProjectPermissionSubmitButton isPending={isPending} />
                </div>
            </AppConfigAvailabilityForm>
        </div>
    );
}
