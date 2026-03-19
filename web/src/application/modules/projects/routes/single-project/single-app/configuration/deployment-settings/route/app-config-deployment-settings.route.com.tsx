import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppDeploymentSettingsCommands, AppDeploymentSettingsQueries } from "~/projects/data";
import { type AppDeploymentSettings } from "~/projects/domain";
import { EAppDeploymentMethod } from "~/projects/module-shared/enums";

import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigDeploymentSettingsForm } from "../form";
import { type AppConfigDeploymentSettingsFormSchemaOutput } from "../schemas";
import { type AppConfigDeploymentSettingsFormRef } from "../types";

export function AppConfigDeploymentSettingsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigDeploymentSettingsFormRef>(null);

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading, error, refetch } = AppDeploymentSettingsQueries.useFindOne({
        projectID: projectId,
        appID: appId,
    });

    const { mutate: update, isPending } = AppDeploymentSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Deployment settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update deployment settings");
            }
        },
    });

    function handleSubmit(values: AppConfigDeploymentSettingsFormSchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");
        invariant(data, "data must be defined");

        const currentSettings = data.data;

        const repoCredentials =
            currentSettings.activeMethod === EAppDeploymentMethod.Repo
                ? currentSettings.repoSource.credentials
                : { id: "", name: "", type: "" as never, status: "" as never, updateVer: 0, createdAt: new Date() };

        const repoPushToRegistry =
            currentSettings.activeMethod === EAppDeploymentMethod.Repo
                ? currentSettings.repoSource.pushToRegistry
                : { id: "", name: "", type: "" as never, status: "" as never, updateVer: 0, createdAt: new Date() };

        const imageRegistryAuth =
            currentSettings.activeMethod === EAppDeploymentMethod.Image
                ? currentSettings.registryAuth
                : { id: "", name: "", type: "" as never, status: "" as never, updateVer: 0, createdAt: new Date() };

        const payload: AppDeploymentSettings =
            values.activeMethod === EAppDeploymentMethod.Repo
                ? {
                      activeMethod: values.activeMethod,
                      command: values.command,
                      workingDir: values.workingDir,
                      preDeploymentCommand: values.preDeploymentCommand,
                      postDeploymentCommand: values.postDeploymentCommand,
                      updateVer: currentSettings.updateVer,
                      notification: currentSettings.notification,
                      repoSource: {
                          buildTool: values.repoSource.buildTool,
                          repoType: values.repoSource.repoType,
                          repoUrl: values.repoSource.repoUrl,
                          repoRef: values.repoSource.repoRef,
                          dockerfilePath: values.repoSource.dockerfilePath ?? "",
                          imageName: values.repoSource.imageName ?? "",
                          imageTags: values.repoSource.imageTags ?? "",
                          credentials: repoCredentials,
                          pushToRegistry: repoPushToRegistry,
                      },
                  }
                : {
                      activeMethod: values.activeMethod,
                      command: values.command,
                      workingDir: values.workingDir,
                      preDeploymentCommand: values.preDeploymentCommand,
                      postDeploymentCommand: values.postDeploymentCommand,
                      updateVer: currentSettings.updateVer,
                      notification: currentSettings.notification,
                      image: values.image,
                      registryAuth: imageRegistryAuth,
                  };

        update({
            projectID: projectId,
            appID: appId,
            payload,
            updateVer: currentSettings.updateVer,
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    // if (error) {
    //     return (
    //         <PageError
    //             error={error}
    //             onRetry={refetch}
    //         />
    //     );
    // }

    // invariant(data, "data must be defined");

    // const { data: settings } = data;

    return (
        <div className="flex flex-col gap-4">
            <h2 className="text-xl font-semibold">Deployment Settings</h2>
            <AppConfigDeploymentSettingsForm
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
                        Deploy
                    </Button>
                </div>
            </AppConfigDeploymentSettingsForm>
        </div>
    );
}
