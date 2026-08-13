import { useMemo, useState } from "react";

import { FieldError, Input, Tabs, TabsList, TabsTrigger } from "@components/ui";
import { Textarea } from "@components/ui/textarea";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { useAppDeploymentSettingsApi } from "~/projects/api/hooks/project-apps";
import { QK } from "~/projects/data/constants";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EDockerfileSource } from "~/projects/module-shared/enums";

import { AppLink, Combobox, InfoBlock } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { DOCKERFILE_TEMPLATE_OPTIONS } from "../constants/dockerfile-template-options.constants";
import { PushToRegistrySelect } from "../form-components";
import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";
import { DockerfileContentEditor } from "./dockerfile-content-editor.com";

export function BuildConfigurationFields({ readOnly = false }: Props) {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const queryClient = useQueryClient();
    const { queries } = useAppDeploymentSettingsApi();
    const [isLoadingTemplate, setIsLoadingTemplate] = useState(false);
    const [templateSearch, setTemplateSearch] = useState("");
    const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null);

    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const dockerfileSource = useWatch({ control, name: "repoSource.dockerfile.source" });
    const isAutoSource = dockerfileSource === EDockerfileSource.Auto;
    const isManualSource = !isAutoSource;

    const { field: sourceField } = useController({
        control,
        name: "repoSource.dockerfile.source",
        defaultValue: EDockerfileSource.Manual,
    });

    const {
        field: dockerfilePath,
        fieldState: { invalid: isDockerfilePathInvalid, error: dockerfilePathError },
    } = useController({ control, name: "repoSource.dockerfile.path" });

    const {
        field: dockerfileContent,
        fieldState: { invalid: isDockerfileContentInvalid, error: dockerfileContentError },
    } = useController({ control, name: "repoSource.dockerfile.content" });

    const {
        field: dockerfileScanPath,
        fieldState: { invalid: isDockerfileScanPathInvalid, error: dockerfileScanPathError },
    } = useController({ control, name: "repoSource.dockerfile.scanPath" });

    const {
        field: imageName,
        fieldState: { invalid: isImageNameInvalid, error: imageNameError },
    } = useController({ control, name: "repoSource.imageName" });

    const templateComboboxOptions = useMemo(() => {
        const normalizedSearch = templateSearch.trim().toLowerCase();

        return DOCKERFILE_TEMPLATE_OPTIONS.filter(option => {
            if (!normalizedSearch) {
                return true;
            }

            return option.label.toLowerCase().includes(normalizedSearch);
        }).map(option => ({
            value: { type: option.type },
            label: option.label,
        }));
    }, [templateSearch]);

    async function handleLoadTemplate(type: string) {
        if (readOnly || !projectId || !env || !appId) {
            return;
        }

        setIsLoadingTemplate(true);

        try {
            const response = await queryClient.fetchQuery({
                queryKey: [
                    QK["projects.apps.deployment-settings.$.dockerfile-template"],
                    { projectID: projectId, env, appID: appId, type },
                ],
                queryFn: ({ signal }) =>
                    queries.getDockerfileTemplate({ projectID: projectId, env, appID: appId, type }, signal),
            });

            dockerfileContent.onChange(response.data.template);
        } catch (error) {
            if (error instanceof Error) {
                toast.error(error.message);
            } else {
                toast.error("Failed to load Dockerfile template");
            }
        } finally {
            setIsLoadingTemplate(false);
        }
    }

    return (
        <>
            <InfoBlock title="Dockerfile Source">
                <Tabs
                    value={
                        sourceField.value === EDockerfileSource.Auto
                            ? EDockerfileSource.Auto
                            : EDockerfileSource.Manual
                    }
                    onValueChange={value => {
                        if (readOnly) {
                            return;
                        }

                        sourceField.onChange(value);
                    }}
                >
                    <TabsList>
                        <TabsTrigger
                            value={EDockerfileSource.Manual}
                            disabled={readOnly}
                        >
                            Manual
                        </TabsTrigger>
                        <TabsTrigger
                            value={EDockerfileSource.Auto}
                            disabled={readOnly}
                        >
                            Auto-Generate
                        </TabsTrigger>
                    </TabsList>
                </Tabs>
            </InfoBlock>

            <InfoBlock title="Dockerfile Path">
                <Input
                    {...dockerfilePath}
                    value={dockerfilePath.value ?? ""}
                    onChange={dockerfilePath.onChange}
                    placeholder="path/to/Dockerfile"
                    aria-invalid={isDockerfilePathInvalid}
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    disabled={readOnly}
                />
                <FieldError errors={[dockerfilePathError]} />
            </InfoBlock>

            {isManualSource && (
                <InfoBlock title="Dockerfile Content">
                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start">
                        <Textarea
                            {...dockerfileContent}
                            value={dockerfileContent.value ?? ""}
                            onChange={dockerfileContent.onChange}
                            placeholder="override Dockerfile content..."
                            aria-invalid={isDockerfileContentInvalid}
                            className={cn(PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS, "min-h-0 flex-1")}
                            minRows={4}
                            maxRows={12}
                            disabled={readOnly}
                        />
                        <div className="flex shrink-0 flex-col gap-2">
                            <Combobox
                                options={templateComboboxOptions}
                                value={selectedTemplate}
                                onChange={value => {
                                    if (!value) {
                                        setSelectedTemplate(null);
                                        return;
                                    }

                                    setSelectedTemplate(value);
                                    void handleLoadTemplate(value).finally(() => {
                                        setSelectedTemplate(null);
                                    });
                                }}
                                onSearch={setTemplateSearch}
                                placeholder="Load Template"
                                searchable
                                closeOnSelect
                                loading={isLoadingTemplate}
                                disabled={readOnly || isLoadingTemplate}
                                valueKey="type"
                                emptyText="No templates found"
                                className="w-[180px]"
                            />
                            <DockerfileContentEditor
                                value={dockerfileContent.value ?? ""}
                                onChange={dockerfileContent.onChange}
                                invalid={isDockerfileContentInvalid}
                                error={dockerfileContentError}
                                readOnly={readOnly}
                            />
                        </div>
                    </div>
                    <FieldError errors={[dockerfileContentError]} />
                </InfoBlock>
            )}

            {isAutoSource && (
                <InfoBlock title="Dockerfile Gen Scan Path">
                    <Input
                        {...dockerfileScanPath}
                        value={dockerfileScanPath.value ?? ""}
                        onChange={dockerfileScanPath.onChange}
                        placeholder="scan/path/to/generate/Dockerfile"
                        aria-invalid={isDockerfileScanPathInvalid}
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        disabled={readOnly}
                    />
                    <FieldError errors={[dockerfileScanPathError]} />
                </InfoBlock>
            )}

            <InfoBlock title="Registry Credentials">
                <p className="text-sm">
                    Need to add private registries for this building?{" "}
                    <AppLink.Basic
                        to={ROUTE.projects.single.providerConfiguration.registryAuth.$route(projectId)}
                        className="text-link"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        Click here
                    </AppLink.Basic>
                </p>
            </InfoBlock>

            <div className={cn(dashedBorderBox, "text-sm leading-6")}>
                <strong>Important:</strong> If your cluster has multiple nodes, a container registry must be
                configured. Without pushing to a registry, other worker nodes won&apos;t be able to access the image to
                run your application.
            </div>

            <PushToRegistrySelect readOnly={readOnly} />

            <InfoBlock title="Image Repository Name">
                <Input
                    {...imageName}
                    value={imageName.value ?? ""}
                    onChange={imageName.onChange}
                    placeholder="auto"
                    aria-invalid={isImageNameInvalid}
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    disabled={readOnly}
                />
                <FieldError errors={[imageNameError]} />
            </InfoBlock>
        </>
    );
}

type Props = {
    readOnly?: boolean;
};
