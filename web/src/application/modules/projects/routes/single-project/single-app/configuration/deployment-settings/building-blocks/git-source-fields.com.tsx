import { useMemo, useState } from "react";

import { Button, Checkbox, Field, FieldError, Input } from "@components/ui";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { BranchesDialog } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EDeploymentRepoOption } from "~/projects/module-shared/enums";
import { canUseGitCredentialSelectors, parseGitRepository } from "~/projects/module-shared/utils";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESettingType } from "@application/shared/enums";

import { GitCredentialSelect, GitRepositoryInput, PushToRegistrySelect } from "../form-components";
import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../schemas";

export function GitSourceFields({ readOnly = false }: Props) {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    const gitSubmodulesOptionId = "repo-options-git-submodules";
    const gitLfsOptionId = "repo-options-git-lfs";
    const [isBranchesDialogOpen, setBranchesDialogOpen] = useState(false);

    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const {
        field: repoRef,
        fieldState: { invalid: isRepoRefInvalid, error: repoRefError },
    } = useController({ control, name: "repoSource.repoRef" });

    const credentials = useWatch({ control, name: "repoSource.credentials" });
    const repoUrlValue = useWatch({ control, name: "repoSource.repoUrl" }) as string | undefined;
    const credentialId = credentials?.id;
    const normalizedRepoUrl = repoUrlValue?.trim() ?? "";
    const repository = useMemo(() => parseGitRepository(normalizedRepoUrl), [normalizedRepoUrl]);
    const canShowGitSelectors = !readOnly && Boolean(credentialId) && canUseGitCredentialSelectors(credentials?.type);
    const canShowBranchesButton = canShowGitSelectors && Boolean(normalizedRepoUrl) && Boolean(repository);
    const isGithubAppCredential = credentials?.type === ESettingType.GithubApp;

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
                <Field>
                    <div className="flex flex-wrap items-start gap-3">
                        <Input
                            {...repoRef}
                            value={repoRef.value}
                            onChange={repoRef.onChange}
                            placeholder="main, tags/v1.2.3"
                            aria-invalid={isRepoRefInvalid}
                            className="w-full min-w-[260px] max-w-[600px] flex-1"
                            disabled={readOnly}
                        />
                        {canShowBranchesButton && (
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => {
                                    setBranchesDialogOpen(true);
                                }}
                            >
                                Show Branches
                            </Button>
                        )}
                    </div>
                    <FieldError errors={[repoRefError]} />
                </Field>
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

            {canShowBranchesButton && credentialId && repository && (
                <BranchesDialog
                    open={isBranchesDialogOpen}
                    onOpenChange={setBranchesDialogOpen}
                    projectId={projectId}
                    credentialId={credentialId}
                    repository={repository}
                    isGithubAppCredential={isGithubAppCredential}
                    onSelect={ref => {
                        repoRef.onChange(ref);
                    }}
                />
            )}
        </>
    );
}

type Props = {
    readOnly?: boolean;
};
