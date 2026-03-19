import { FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { EBuildTool, ERepoType } from "~/projects/module-shared/enums";

import { Combobox, InfoBlock } from "@application/shared/components";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

const MOCK_GIT_CREDENTIALS = [
    { value: { id: "cred-1", name: "GitHub Token" }, label: "GitHub Token" },
    { value: { id: "cred-2", name: "GitLab Token" }, label: "GitLab Token" },
];

const MOCK_REGISTRY_CREDENTIALS = [
    { value: { id: "reg-1", name: "Docker Hub" }, label: "Docker Hub" },
    { value: { id: "reg-2", name: "GitHub Container Registry" }, label: "GitHub Container Registry" },
];

const BUILD_TOOL_OPTIONS = [
    { value: { id: EBuildTool.Docker as string }, label: "Docker" },
    { value: { id: EBuildTool.Nixpacks as string }, label: "Nixpacks" },
];

const REPO_TYPE_OPTIONS = [
    { value: { id: ERepoType.Git }, label: "Git" },
];

export function GitSourceFields() {
    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const {
        field: credentials,
        fieldState: { invalid: isCredentialsInvalid, error: credentialsError },
    } = useController({ control, name: "repoSource.credentials" });

    const {
        field: repoUrl,
        fieldState: { invalid: isRepoUrlInvalid, error: repoUrlError },
    } = useController({ control, name: "repoSource.repoUrl" });

    const {
        field: repoRef,
        fieldState: { invalid: isRepoRefInvalid, error: repoRefError },
    } = useController({ control, name: "repoSource.repoRef" });

    const {
        field: dockerfilePath,
        fieldState: { invalid: isDockerfilePathInvalid, error: dockerfilePathError },
    } = useController({ control, name: "repoSource.dockerfilePath" });

    const {
        field: pushToRegistry,
        fieldState: { invalid: isPushToRegistryInvalid, error: pushToRegistryError },
    } = useController({ control, name: "repoSource.pushToRegistry" });

    const {
        field: imageName,
        fieldState: { invalid: isImageNameInvalid, error: imageNameError },
    } = useController({ control, name: "repoSource.imageName" });

    const {
        field: buildTool,
        fieldState: { invalid: isBuildToolInvalid, error: buildToolError },
    } = useController({ control, name: "repoSource.buildTool" });

    const {
        field: repoType,
        fieldState: { invalid: isRepoTypeInvalid, error: repoTypeError },
    } = useController({ control, name: "repoSource.repoType" });

    return (
        <>
            <InfoBlock title="Git Credential">
                <Combobox
                    options={MOCK_GIT_CREDENTIALS}
                    value={credentials.value?.id ?? null}
                    onChange={(_, option) => {
                        credentials.onChange(option ?? null);
                    }}
                    placeholder="Select git credential"
                    searchable={false}
                    closeOnSelect
                    className="max-w-[400px]"
                    valueKey="id"
                    aria-invalid={isCredentialsInvalid}
                />
                <FieldError errors={[credentialsError]} />
            </InfoBlock>

            <InfoBlock
                title={
                    <>
                        Git Repository <span className="text-destructive">*</span>
                    </>
                }
            >
                <Input
                    {...repoUrl}
                    value={repoUrl.value}
                    onChange={repoUrl.onChange}
                    placeholder="https://domain/path/repo"
                    aria-invalid={isRepoUrlInvalid}
                    className="max-w-[400px]"
                />
                <FieldError errors={[repoUrlError]} />
            </InfoBlock>

            <InfoBlock
                title={
                    <>
                        Branch / Commit <span className="text-destructive">*</span>
                    </>
                }
            >
                <Input
                    {...repoRef}
                    value={repoRef.value}
                    onChange={repoRef.onChange}
                    placeholder="main, tags/v1.2.3"
                    aria-invalid={isRepoRefInvalid}
                    className="max-w-[400px]"
                />
                <FieldError errors={[repoRefError]} />
            </InfoBlock>

            <InfoBlock title="Dockerfile Path">
                <Input
                    {...dockerfilePath}
                    value={dockerfilePath.value ?? ""}
                    onChange={dockerfilePath.onChange}
                    placeholder="path/to/Dockerfile"
                    aria-invalid={isDockerfilePathInvalid}
                    className="max-w-[400px]"
                />
                <FieldError errors={[dockerfilePathError]} />
            </InfoBlock>

            <InfoBlock title="Registry Credentials">
                <Combobox
                    options={MOCK_REGISTRY_CREDENTIALS}
                    value={pushToRegistry.value?.id ?? null}
                    onChange={(_, option) => {
                        pushToRegistry.onChange(option ?? null);
                    }}
                    placeholder="Select registry credentials"
                    searchable={false}
                    closeOnSelect
                    className="max-w-[400px]"
                    valueKey="id"
                    aria-invalid={isPushToRegistryInvalid}
                />
                <FieldError errors={[pushToRegistryError]} />
            </InfoBlock>

            <InfoBlock title="Image Repository Name">
                <Input
                    {...imageName}
                    value={imageName.value ?? ""}
                    onChange={imageName.onChange}
                    placeholder="auto"
                    aria-invalid={isImageNameInvalid}
                    className="max-w-[400px]"
                />
                <FieldError errors={[imageNameError]} />
            </InfoBlock>

            <InfoBlock title="Build Tool">
                <Combobox
                    options={BUILD_TOOL_OPTIONS}
                    value={buildTool.value}
                    onChange={value => {
                        buildTool.onChange(value);
                    }}
                    placeholder="Select build tool"
                    searchable={false}
                    closeOnSelect
                    className="max-w-[400px]"
                    valueKey="id"
                    aria-invalid={isBuildToolInvalid}
                />
                <FieldError errors={[buildToolError]} />
            </InfoBlock>

            <InfoBlock title="Repository Type">
                <Combobox
                    options={REPO_TYPE_OPTIONS}
                    value={repoType.value}
                    onChange={value => {
                        repoType.onChange(value);
                    }}
                    placeholder="Select repo type"
                    searchable={false}
                    closeOnSelect
                    className="max-w-[400px]"
                    valueKey="id"
                    aria-invalid={isRepoTypeInvalid}
                />
                <FieldError errors={[repoTypeError]} />
            </InfoBlock>
        </>
    );
}
