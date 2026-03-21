import { FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { EBuildTool, ERepoType } from "~/projects/module-shared/enums";

import { Combobox, InfoBlock } from "@application/shared/components";

import {
    GitCredentialSelect,
    GitRepositoryInput,
    PushToRegistrySelect,
} from "../form-components";
import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

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
        field: repoRef,
        fieldState: { invalid: isRepoRefInvalid, error: repoRefError },
    } = useController({ control, name: "repoSource.repoRef" });

    const {
        field: dockerfilePath,
        fieldState: { invalid: isDockerfilePathInvalid, error: dockerfilePathError },
    } = useController({ control, name: "repoSource.dockerfilePath" });

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
            <GitCredentialSelect />

            <GitRepositoryInput />

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

            <PushToRegistrySelect />

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
