import { useRef } from "react";

import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppNetworkSettingsCommands, AppNetworkSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { type AppNetworkSettings } from "~/projects/domain";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigNetworksForm } from "../form";
import { type AppConfigNetworksFormSchemaOutput } from "../schemas";
import { type AppConfigNetworksFormRef } from "../types";

function splitBySpace(value: string): string[] {
    return value.trim().split(/\s+/).filter(Boolean);
}

function mapFormValuesToPayload(
    values: AppConfigNetworksFormSchemaOutput,
    server: AppNetworkSettings | undefined,
): AppNetworkSettings {
    return {
        networkAttachments: values.networkAttachments.map(item => ({
            id: item.id.trim(),
            aliases: splitBySpace(item.aliasesText),
        })),
        hostsFileEntries: values.hostsFileEntries.map(item => ({
            address: item.address.trim(),
            hostnames: splitBySpace(item.hostnamesText),
        })),
        dnsConfig: {
            nameservers: values.dnsConfig.nameservers.map(item => item.value.trim()).filter(Boolean),
            search: values.dnsConfig.search.map(item => item.value.trim()).filter(Boolean),
            options: values.dnsConfig.options.map(item => item.value.trim()).filter(Boolean),
        },
        endpointSpec: {
            mode: values.resolutionMode,
            ports: values.portConfigs.map(item => ({
                published: item.published,
                target: item.target,
                protocol: item.protocol,
                publishMode: item.publishMode,
            })),
        },
        updateVer: server?.updateVer ?? 0,
    };
}

export function AppConfigNetworksRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigNetworksFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppNetworkSettingsQueries.useFindOne(
        {
            projectID: projectId,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const { mutate: update, isPending } = AppNetworkSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Network settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update network settings");
            }
        },
    });

    function handleSubmit(values: AppConfigNetworksFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

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
                    href="/cluster/networks"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-500 underline underline-offset-2"
                >
                    docs
                </a>
            </div>

            <AppConfigNetworksForm
                ref={formRef}
                defaultValues={data?.data}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
            >
                <div className="flex justify-end mt-4">
                    <ProjectPermissionSubmitButton isPending={isPending} />
                </div>
            </AppConfigNetworksForm>
        </div>
    );
}
