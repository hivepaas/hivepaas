import { useState } from "react";

import { Button, Field, FieldError, Input } from "@components/ui";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { GitRepositoriesDialog } from "~/projects/module-shared/components";
import { canUseGitCredentialSelectors } from "~/projects/module-shared/utils";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

export function GitRepositoryInput({ readOnly = false }: Props) {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");
    const [isRepositoriesDialogOpen, setRepositoriesDialogOpen] = useState(false);

    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const credentials = useWatch({ control, name: "repoSource.credentials" });
    const credentialId = credentials?.id;

    const {
        field: repoUrl,
        fieldState: { invalid: isRepoUrlInvalid, error: repoUrlError },
    } = useController({ control, name: "repoSource.repoUrl" });

    const canShowRepositoriesButton =
        !readOnly && Boolean(credentialId) && canUseGitCredentialSelectors(credentials?.type);

    return (
        <>
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Git Repository"
                        isRequired
                    />
                }
            >
                <Field>
                    <div className="flex flex-wrap items-start gap-3">
                        <Input
                            {...repoUrl}
                            value={repoUrl.value}
                            onChange={repoUrl.onChange}
                            placeholder="Paste repository URL"
                            aria-invalid={isRepoUrlInvalid}
                            className="w-full min-w-[260px] max-w-[600px] flex-1"
                            disabled={readOnly}
                        />
                        {canShowRepositoriesButton && credentialId && (
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => {
                                    setRepositoriesDialogOpen(true);
                                }}
                            >
                                Show Repos
                            </Button>
                        )}
                    </div>
                    <FieldError errors={[repoUrlError]} />
                </Field>
            </InfoBlock>

            {canShowRepositoriesButton && credentialId && (
                <GitRepositoriesDialog
                    open={isRepositoriesDialogOpen}
                    onOpenChange={setRepositoriesDialogOpen}
                    projectId={projectId}
                    credentialId={credentialId}
                    onSelect={repository => {
                        repoUrl.onChange(repository.cloneURL);
                    }}
                />
            )}
        </>
    );
}

type Props = {
    readOnly?: boolean;
};
