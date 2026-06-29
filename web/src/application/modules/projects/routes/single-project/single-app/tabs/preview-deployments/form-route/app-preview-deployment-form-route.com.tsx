import { useEffect, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { toast } from "sonner";
import type { AppPreviews_PrepareCreate_Res } from "~/projects/api/services";
import { AppDeploymentSettingsQueries, AppPreviewsCommands } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { EAppDeploymentMethod } from "~/projects/module-shared/enums";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { ESettingType } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

import { AppPreviewDeploymentForm } from "../form";
import type { AppPreviewDeploymentFormOutput } from "../schemas";

type PreparedPreview = AppPreviews_PrepareCreate_Res["data"];

export function AppPreviewDeploymentFormRoute({ projectId, appId, initialPreparedPreview }: Props) {
    const { navigate } = useAppNavigate();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const [preparedPreview, setPreparedPreview] = useState<PreparedPreview | undefined>(initialPreparedPreview);
    const hasRequestedPrepareRef = useRef(Boolean(initialPreparedPreview));
    const preparedCredentialId = preparedPreview?.repoCredentials?.id;

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.apps.single.previewDeployments.$route(projectId, appId), {
            ignorePrevPath: true,
        });
    }

    const { mutate: preparePreview, isPending: isPreparing } = AppPreviewsCommands.usePrepareCreate({
        onSuccess: response => {
            setPreparedPreview(response.data);
        },
        onError: () => {
            navigateToList();
        },
    });

    const { mutate: createPreview, isPending: isCreating } = AppPreviewsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Preview deployment created successfully");
            navigateToList();
        },
    });

    const deploymentSettingsQuery = AppDeploymentSettingsQueries.useFindOne(
        {
            projectID: projectId,
            appID: appId,
        },
        {
            ...APP_CONFIGURATION_QUERY_OPTIONS,
            enabled: Boolean(preparedCredentialId),
        },
    );
    const deploymentSettings = deploymentSettingsQuery.data?.data;
    const deploymentGitCredential =
        deploymentSettings?.activeMethod === EAppDeploymentMethod.Repo
            ? deploymentSettings.repoSource.credentials
            : null;
    const isGitCredentialTypeResolved = !preparedCredentialId || deploymentGitCredential?.id === preparedCredentialId;
    const isGithubAppCredential = Boolean(
        preparedCredentialId &&
            deploymentGitCredential?.id === preparedCredentialId &&
            deploymentGitCredential.type === ESettingType.GithubApp,
    );

    useEffect(() => {
        if (preparedPreview || hasRequestedPrepareRef.current) {
            return;
        }

        hasRequestedPrepareRef.current = true;
        preparePreview({
            projectID: projectId,
            appID: appId,
        });
    }, [appId, preparePreview, preparedPreview, projectId]);

    function onSubmit(values: AppPreviewDeploymentFormOutput) {
        if (!canWrite) {
            return;
        }

        createPreview({
            projectID: projectId,
            appID: appId,
            repoRef: values.repoRef,
            customSubdomain: values.customSubdomain,
            noStart: values.noStart,
        });
    }

    return (
        <section className={cn(listBox)}>
            <RouteFormHeader title="Create a preview" />

            {!preparedPreview && isPreparing && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {preparedPreview && (
                <AppPreviewDeploymentForm
                    projectId={projectId}
                    appId={appId}
                    preparedPreview={preparedPreview}
                    isGitCredentialTypeResolved={isGitCredentialTypeResolved}
                    isGithubAppCredential={isGithubAppCredential}
                    isPending={isCreating}
                    readOnly={!canWrite}
                    onSubmit={onSubmit}
                />
            )}
        </section>
    );
}

interface Props {
    projectId: string;
    appId: string;
    initialPreparedPreview?: PreparedPreview;
}
