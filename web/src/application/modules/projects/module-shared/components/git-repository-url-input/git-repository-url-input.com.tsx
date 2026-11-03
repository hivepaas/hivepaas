import { useState } from "react";

import { canUseGitCredentialSelectors } from "~/projects/module-shared/utils";

import { Button, Input } from "@/components/ui";

import { GitRepositoriesDialog } from "../git-selectors";

/**
 * A repository's URL, typed or picked: Show Repos lists the repositories the
 * chosen credentials reach, as the deployment settings do, for the credentials
 * that can list them.
 */
export function GitRepositoryUrlInput({
    projectId,
    env,
    credentials,
    value,
    onChange,
    readOnly = false,
    invalid = false,
    className,
}: GitRepositoryUrlInputProps) {
    const [isRepositoriesDialogOpen, setRepositoriesDialogOpen] = useState(false);
    const credentialId = credentials?.id ?? "";
    const canShowRepositories = !readOnly && credentialId !== "" && canUseGitCredentialSelectors(credentials?.type);

    return (
        <>
            <div className="flex w-full items-center gap-2">
                <Input
                    value={value}
                    onChange={event => {
                        onChange(event.target.value);
                    }}
                    placeholder="https://github.com/owner/repo.git"
                    aria-invalid={invalid}
                    className={className ?? "min-w-0 flex-1"}
                    disabled={readOnly}
                />
                {canShowRepositories && (
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

            {canShowRepositories && (
                <GitRepositoriesDialog
                    open={isRepositoriesDialogOpen}
                    onOpenChange={setRepositoriesDialogOpen}
                    projectId={projectId}
                    env={env}
                    credentialId={credentialId}
                    onSelect={repository => {
                        onChange(repository.cloneURL);
                    }}
                />
            )}
        </>
    );
}

export interface GitRepositoryUrlInputProps {
    projectId: string;
    env: string;
    /** The credentials the repositories are listed with; none, or a kind that cannot list, hides the button. */
    credentials: { id: string; type?: string } | null | undefined;
    value: string;
    onChange: (value: string) => void;
    readOnly?: boolean;
    invalid?: boolean;
    className?: string;
}
