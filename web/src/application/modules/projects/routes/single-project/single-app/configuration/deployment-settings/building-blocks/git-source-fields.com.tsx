import { Checkbox, FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EDeploymentRepoOption } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { GitCredentialSelect, GitRepositoryInput, PushToRegistrySelect } from "../form-components";
import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

export function GitSourceFields({ readOnly = false }: Props) {
    const gitSubmodulesOptionId = "repo-options-git-submodules";
    const gitLfsOptionId = "repo-options-git-lfs";

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
        field: commitHash,
        fieldState: { invalid: isCommitHashInvalid, error: commitHashError },
    } = useController({ control, name: "repoSource.commitHash" });

    const { field: gitSubmodulesEnabled } = useController({
        control,
        name: `repoSource.repoOptions.${EDeploymentRepoOption.GitSubmodulesEnabled}`,
        defaultValue: true,
    });

    const { field: gitLfsEnabled } = useController({
        control,
        name: `repoSource.repoOptions.${EDeploymentRepoOption.GitLfsEnabled}`,
        defaultValue: true,
    });

    const {
        field: dockerfilePath,
        fieldState: { invalid: isDockerfilePathInvalid, error: dockerfilePathError },
    } = useController({ control, name: "repoSource.dockerfilePath" });

    const {
        field: imageName,
        fieldState: { invalid: isImageNameInvalid, error: imageNameError },
    } = useController({ control, name: "repoSource.imageName" });

    return (
        <>
            <GitCredentialSelect readOnly={readOnly} />

            <GitRepositoryInput readOnly={readOnly} />

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Branch"
                        isRequired
                    />
                }
            >
                <Input
                    {...repoRef}
                    value={repoRef.value}
                    onChange={repoRef.onChange}
                    placeholder="main, tags/v1.2.3"
                    aria-invalid={isRepoRefInvalid}
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    disabled={readOnly}
                />
                <FieldError errors={[repoRefError]} />
            </InfoBlock>

            <InfoBlock title="Commit Hash">
                <Input
                    {...commitHash}
                    value={commitHash.value ?? ""}
                    onChange={commitHash.onChange}
                    placeholder="commit sha"
                    aria-invalid={isCommitHashInvalid}
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    disabled={readOnly}
                />
                <FieldError errors={[commitHashError]} />
            </InfoBlock>

            <InfoBlock title="Repository Options">
                <div className="flex flex-wrap items-center gap-8">
                    <div className="flex items-center gap-3">
                        <Checkbox
                            id={gitSubmodulesOptionId}
                            checked={gitSubmodulesEnabled.value}
                            onCheckedChange={checked => {
                                gitSubmodulesEnabled.onChange(checked === true);
                            }}
                            disabled={readOnly}
                        />
                        <label
                            htmlFor={gitSubmodulesOptionId}
                            className="text-sm"
                        >
                            Git Submodules
                        </label>
                    </div>
                    <div className="flex items-center gap-3">
                        <Checkbox
                            id={gitLfsOptionId}
                            checked={gitLfsEnabled.value}
                            onCheckedChange={checked => {
                                gitLfsEnabled.onChange(checked === true);
                            }}
                            disabled={readOnly}
                        />
                        <label
                            htmlFor={gitLfsOptionId}
                            className="text-sm"
                        >
                            Git LFS
                        </label>
                    </div>
                </div>
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

            {/* <InfoBlock title="Build Tool">
                <Combobox
                    options={BUILD_TOOL_OPTIONS}
                    value={buildTool.value}
                    onChange={value => {
                        buildTool.onChange(value);
                    }}
                    placeholder="Select build tool"
                    searchable={false}
                    closeOnSelect
                    className="max-w-[600px]"
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
                    className="max-w-[600px]"
                    valueKey="id"
                    aria-invalid={isRepoTypeInvalid}
                />
                <FieldError errors={[repoTypeError]} />
            </InfoBlock> */}
        </>
    );
}

type Props = {
    readOnly?: boolean;
};
