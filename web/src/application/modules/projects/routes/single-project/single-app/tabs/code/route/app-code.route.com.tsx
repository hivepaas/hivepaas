import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppDeploymentSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { EAppDeploymentMethod } from "~/projects/module-shared/enums";
import { withLockFiles } from "~/projects/module-shared/utils";

import { AppLink, AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { FunctionCodeWorkspace, FunctionTestPanel } from "../building-blocks";

/**
 * A function's code: its files in the editor, saved and deployed together.
 */
export function AppCodeRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { data, isLoading } = AppDeploymentSettingsQueries.useFindOne(
        { projectID: projectId, env, appID: appId },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );
    const settings = data?.data;

    if (isLoading || !settings) {
        return <AppLoader />;
    }

    if (settings.activeMethod !== EAppDeploymentMethod.Function) {
        return (
            <div className={cn(listBox, "text-sm text-muted-foreground")}>
                This app is not a function: its code is not kept here.
            </div>
        );
    }

    const { repo } = settings.functionSource.code;
    if (repo) {
        return (
            <div className={cn(listBox, "flex flex-col gap-1 text-sm")}>
                <span>The function&apos;s code is in a repository, where it is edited:</span>
                <span className="font-mono">
                    {repo.repoUrl} {repo.repoRef}
                    {settings.functionSource.code.dir ? ` - ${settings.functionSource.code.dir}` : ""}
                </span>
                <AppLink.Basic
                    to={ROUTE.projects.single.apps.single.configuration.deploymentSettings.$route(
                        projectId,
                        env,
                        appId,
                    )}
                    className="text-link"
                >
                    Change the repository in the function&apos;s settings
                </AppLink.Basic>
            </div>
        );
    }

    // The editor and the test panel side by side want more room than a list.
    return (
        <div className={cn(listBox, "max-w-[1800px]")}>
            <FunctionCodeWorkspace
                key={settings.updateVer}
                projectId={projectId}
                env={env}
                appId={appId}
                settings={settings}
                readOnly={!canWrite}
            >
                {(files, setFiles) => (
                    <FunctionTestPanel
                        projectId={projectId}
                        env={env}
                        appId={appId}
                        files={files}
                        readOnly={!canWrite}
                        onAddFiles={lockFiles => {
                            setFiles(withLockFiles(files, lockFiles));
                        }}
                    />
                )}
            </FunctionCodeWorkspace>
        </div>
    );
}
