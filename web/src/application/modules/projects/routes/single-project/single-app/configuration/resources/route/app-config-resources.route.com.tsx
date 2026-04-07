import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppResourceSettingsCommands, AppResourceSettingsQueries } from "~/projects/data";
import { type AppResourceSettings } from "~/projects/domain";

import { AppLoader } from "@application/shared/components";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigResourcesForm } from "../form";
import { type AppConfigResourcesFormSchemaOutput } from "../schemas";
import { type AppConfigResourcesFormRef } from "../types";

function mapFormValuesToPayload(
    values: AppConfigResourcesFormSchemaOutput,
    server: AppResourceSettings | undefined,
): AppResourceSettings {
    const sysctls: Record<string, string> = {};
    for (const row of values.capabilities.sysctls) {
        sysctls[row.name] = row.value;
    }

    return {
        reservations:
            values.reservations.cpus != null ||
            values.reservations.memoryMB != null ||
            values.reservations.genericResources.length > 0
                ? {
                      cpus: values.reservations.cpus,
                      memoryMB: values.reservations.memoryMB,
                      genericResources: values.reservations.genericResources.map(item => ({
                          kind: item.kind,
                          value: item.value,
                      })),
                  }
                : null,
        limits:
            values.limits.cpus != null || values.limits.memoryMB != null || values.limits.pids != null
                ? {
                      cpus: values.limits.cpus,
                      memoryMB: values.limits.memoryMB,
                      pids: values.limits.pids,
                  }
                : null,
        ulimits: values.ulimits.map(item => ({
            name: item.name,
            hard: item.hard,
            soft: item.soft,
        })),
        capabilities: {
            capabilityAdd: values.capabilities.capabilityAdd ? values.capabilities.capabilityAdd.split(" ") : [],
            capabilityDrop: values.capabilities.capabilityDrop ? values.capabilities.capabilityDrop.split(" ") : [],
            enableGPU: values.capabilities.enableGPU,
            oomScoreAdj: values.capabilities.oomScoreAdj,
            sysctls,
        },
        updateVer: server?.updateVer ?? 0,
    };
}

export function AppConfigResourcesRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigResourcesFormRef>(null);

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppResourceSettingsQueries.useFindOne({
        projectID: projectId,
        appID: appId,
    });

    const { mutate: update, isPending } = AppResourceSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Resource settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update resource settings");
            }
        },
    });

    function handleSubmit(values: AppConfigResourcesFormSchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");

        update({
            projectID: projectId,
            appID: appId,
            payload: mapFormValuesToPayload(values, data?.data),
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <div className="rounded-md border border-dashed border-primary bg-accent p-2 text-sm text-muted-foreground">
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

            <AppConfigResourcesForm
                ref={formRef}
                defaultValues={data?.data}
                onSubmit={handleSubmit}
            >
                <div className="flex justify-end mt-4">
                    <Button
                        type="submit"
                        className="min-w-[100px]"
                        disabled={isPending}
                        isLoading={isPending}
                    >
                        Save
                    </Button>
                </div>
            </AppConfigResourcesForm>
        </div>
    );
}
