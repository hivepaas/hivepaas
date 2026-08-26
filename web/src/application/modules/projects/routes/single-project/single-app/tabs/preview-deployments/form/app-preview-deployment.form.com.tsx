import { useMemo, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import type { AppPreviews_PrepareCreate_Res } from "~/projects/api/services";
import { BranchesDialog, PullRequestsDialog } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { parseGitRepository } from "~/projects/module-shared/utils";

import { AppLink, InfoBlock } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import {
    type AppPreviewDeploymentFormInput,
    type AppPreviewDeploymentFormOutput,
    AppPreviewDeploymentFormSchema,
    DEFAULT_APP_PREVIEW_DEPLOYMENT_FORM_VALUES,
    PREVIEW_DEPLOYMENT_TRIGGER,
    type PreviewDeploymentTrigger,
} from "../schemas";

const INFO_BLOCK_TITLE_WIDTH = 300;

type PreparedPreview = AppPreviews_PrepareCreate_Res["data"];

export function AppPreviewDeploymentForm({
    projectId,
    env,
    appId,
    preparedPreview,
    isGitCredentialTypeResolved,
    isGithubAppCredential,
    isPending,
    readOnly = false,
    onSubmit,
}: Props) {
    const isCloneDbAppsAllowed = preparedPreview.canCloneDbApps || preparedPreview.canSkipCloningDbApps;
    const [isPullRequestsDialogOpen, setPullRequestsDialogOpen] = useState(false);
    const [isBranchesDialogOpen, setBranchesDialogOpen] = useState(false);
    const {
        handleSubmit,
        control,
        setValue,
        formState: { errors },
    } = useForm<AppPreviewDeploymentFormInput, unknown, AppPreviewDeploymentFormOutput>({
        defaultValues: {
            ...DEFAULT_APP_PREVIEW_DEPLOYMENT_FORM_VALUES,
            cloneDbApps: isCloneDbAppsAllowed,
        },
        resolver: zodResolver(AppPreviewDeploymentFormSchema),
        mode: "onSubmit",
    });

    const { field: trigger } = useController({ control, name: "trigger" });
    const {
        field: repoRef,
        fieldState: { invalid: isRepoRefInvalid },
    } = useController({ control, name: "repoRef" });
    const {
        field: customSubdomain,
        fieldState: { invalid: isCustomSubdomainInvalid },
    } = useController({ control, name: "customSubdomain" });
    const { field: cloneDbApps } = useController({ control, name: "cloneDbApps" });
    const { field: noStart } = useController({ control, name: "noStart" });

    const repository = useMemo(() => parseGitRepository(preparedPreview.repoURL), [preparedPreview.repoURL]);
    const credentialId = preparedPreview.repoCredentials?.id;
    const canUseSelectors = Boolean(credentialId && repository);
    const canShowPullRequestsButton = preparedPreview.canListPullRequests && canUseSelectors;
    const canShowBranchesButton = preparedPreview.canListBranches && canUseSelectors;
    const isSelectorDisabled = readOnly || !isGitCredentialTypeResolved;
    const isDashboardUI = trigger.value === PREVIEW_DEPLOYMENT_TRIGGER.DashboardUI;

    function handleSelectGitRef(ref: string) {
        setValue("repoRef", ref, {
            shouldDirty: true,
            shouldValidate: true,
        });
    }

    function onValid(values: AppPreviewDeploymentFormOutput) {
        if (readOnly || values.trigger !== PREVIEW_DEPLOYMENT_TRIGGER.DashboardUI) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<AppPreviewDeploymentFormOutput>) {
        console.error(_errors);
    }

    return (
        <>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    if (readOnly) {
                        return;
                    }

                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="flex flex-col gap-7"
            >
                <FieldGroup className="gap-6">
                    <InfoBlock
                        title="Preview Deployment Trigger"
                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                    >
                        <Tabs
                            value={trigger.value}
                            onValueChange={value => {
                                trigger.onChange(value as PreviewDeploymentTrigger);
                            }}
                        >
                            <TabsList>
                                <TabsTrigger value={PREVIEW_DEPLOYMENT_TRIGGER.PullRequestComments}>
                                    Pull Request Comments
                                </TabsTrigger>
                                <TabsTrigger value={PREVIEW_DEPLOYMENT_TRIGGER.DashboardUI}>Dashboard UI</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </InfoBlock>

                    {!isDashboardUI && (
                        <div className={cn(dashedBorderBox, "space-y-4 text-sm leading-6")}>
                            <div className="space-y-2">
                                <p>Comment in your pull request to trigger a preview deployment:</p>
                                <div className="rounded bg-muted/60 p-3 font-mono text-sm text-foreground">
                                    /hivepaas deploy [subdomain=&lt;name&gt;] [clonedb|noclonedb] [nowait] [nostart]
                                </div>
                                <div className="space-y-1.5 pt-1 text-sm text-muted-foreground">
                                    <p className="font-semibold text-foreground">Available flags:</p>
                                    <ul className="list-disc space-y-1 pl-5">
                                        <li>
                                            <code className="text-foreground">subdomain=&lt;name&gt;</code>: Set a
                                            custom subdomain (default:{" "}
                                            <code className="text-foreground">pr-&lt;number&gt;</code>).
                                        </li>
                                        <li>
                                            <code className="text-foreground">clonedb</code> /{" "}
                                            <code className="text-foreground">noclonedb</code>: Force clone or skip
                                            cloning configured database apps.
                                        </li>
                                        <li>
                                            <code className="text-foreground">nowait</code>: Deploy immediately,
                                            ignoring the configured creation delay.
                                        </li>
                                        <li>
                                            <code className="text-foreground">nostart</code>: Create the preview app
                                            without starting containers.
                                        </li>
                                    </ul>
                                </div>
                            </div>

                            <div className="space-y-2">
                                <p>To cancel an in-progress preview deployment:</p>
                                <div className="rounded bg-muted/60 p-3 font-mono text-sm text-foreground">
                                    /hivepaas cancel
                                </div>
                            </div>

                            <p className="border-t pt-3 text-sm text-muted-foreground">
                                <span className="font-semibold text-orange-500">Note:</span> This method only works if
                                you have configured the GitHub App or webhook correctly, and preview deployments have
                                been enabled{" "}
                                <AppLink.Modules
                                    to={ROUTE.projects.single.apps.single.configuration.deploymentSettings.$route(
                                        projectId,
                                        env,
                                        appId,
                                    )}
                                    className="font-medium text-primary underline-offset-4 hover:underline"
                                >
                                    here
                                </AppLink.Modules>
                                .
                            </p>
                        </div>
                    )}

                    {isDashboardUI && (
                        <>
                            <InfoBlock
                                title="Repository"
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                            >
                                <div className="text-sm leading-9 text-foreground">{preparedPreview.repoURL}</div>
                            </InfoBlock>

                            <InfoBlock
                                title="Git Ref"
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                            >
                                <Field>
                                    <div className="flex flex-wrap items-start gap-3">
                                        <Input
                                            {...repoRef}
                                            placeholder="pull/123 (github, gitea), merge-requests/123 (gitlab)"
                                            aria-invalid={isRepoRefInvalid}
                                            disabled={readOnly}
                                            className="w-full min-w-[260px] max-w-[600px] flex-1"
                                        />
                                        {canShowPullRequestsButton && repository && credentialId && (
                                            <Button
                                                type="button"
                                                variant="outline"
                                                disabled={isSelectorDisabled}
                                                onClick={() => {
                                                    setPullRequestsDialogOpen(true);
                                                }}
                                            >
                                                Show PRs
                                            </Button>
                                        )}
                                        {canShowBranchesButton && repository && credentialId && (
                                            <Button
                                                type="button"
                                                variant="outline"
                                                disabled={isSelectorDisabled}
                                                onClick={() => {
                                                    setBranchesDialogOpen(true);
                                                }}
                                            >
                                                Show Branches
                                            </Button>
                                        )}
                                    </div>
                                    <FieldError errors={[errors.repoRef]} />
                                </Field>
                            </InfoBlock>

                            <InfoBlock
                                title="Custom Subdomain"
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                            >
                                <Field>
                                    <Input
                                        {...customSubdomain}
                                        placeholder="auto"
                                        aria-invalid={isCustomSubdomainInvalid}
                                        disabled={readOnly}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    />
                                    <FieldError errors={[errors.customSubdomain]} />
                                </Field>
                            </InfoBlock>

                            <InfoBlock
                                title="Clone DB Apps"
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                            >
                                <Checkbox
                                    checked={Boolean(cloneDbApps.value)}
                                    disabled={readOnly || !isCloneDbAppsAllowed}
                                    onCheckedChange={checked => {
                                        cloneDbApps.onChange(checked === true);
                                    }}
                                />
                            </InfoBlock>

                            <InfoBlock
                                title="Create Only (Do Not Start)"
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                            >
                                <Checkbox
                                    checked={noStart.value}
                                    disabled={readOnly}
                                    onCheckedChange={checked => {
                                        noStart.onChange(checked === true);
                                    }}
                                />
                            </InfoBlock>

                            <div className="flex justify-end pt-8">
                                <Button
                                    type="submit"
                                    isLoading={isPending}
                                    disabled={readOnly}
                                >
                                    Deploy
                                </Button>
                            </div>
                        </>
                    )}
                </FieldGroup>
            </form>

            {repository && credentialId && (
                <>
                    <PullRequestsDialog
                        open={isPullRequestsDialogOpen}
                        onOpenChange={setPullRequestsDialogOpen}
                        projectId={projectId}
                        env={env}
                        credentialId={credentialId}
                        repository={repository}
                        isGithubAppCredential={isGithubAppCredential}
                        onSelect={handleSelectGitRef}
                    />
                    <BranchesDialog
                        open={isBranchesDialogOpen}
                        onOpenChange={setBranchesDialogOpen}
                        projectId={projectId}
                        env={env}
                        credentialId={credentialId}
                        repository={repository}
                        isGithubAppCredential={isGithubAppCredential}
                        onSelect={handleSelectGitRef}
                        selectLabel="Deploy"
                        dialogWidthClassName="w-[1000px]"
                    />
                </>
            )}
        </>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    preparedPreview: PreparedPreview;
    isGitCredentialTypeResolved: boolean;
    isGithubAppCredential: boolean;
    isPending: boolean;
    readOnly?: boolean;
    onSubmit: (values: AppPreviewDeploymentFormOutput) => void;
}
