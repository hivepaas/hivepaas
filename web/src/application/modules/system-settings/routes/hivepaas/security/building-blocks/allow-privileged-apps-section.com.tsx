import { ServerCog } from "lucide-react";
import { HivePaaSSecuritySettingsQueries } from "~/system-settings/data";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { SecuritySwitchSection } from "./security-switch-section.com";

export function AllowPrivilegedAppsSection() {
    return (
        <SecuritySwitchSection
            field="allowPrivilegedApps"
            title="Privileged Apps"
            action="Privileged Apps"
            description={
                <div className="flex flex-col gap-2">
                    <p>
                        Apps that start containers - CI runners, database cluster managers - do not need this: HivePaaS
                        gives them the Docker API through a proxy that lets them do only what their Docker API settings
                        say.
                    </p>
                    <p>
                        This switch is for what the proxy cannot serve. With it on, an administrator can give an app the
                        node&apos;s own Docker socket, and an import can mount a directory of the node into an app.
                    </p>
                    <p className="text-amber-700 dark:text-amber-400">
                        Such an app has full control of its node, and on a manager node of the whole cluster, including
                        HivePaaS itself and the data of every other app. Keep this disabled unless you need one. Turning
                        it off gives no new app this access, and takes it from none of the apps listed below.
                    </p>
                    <PrivilegedAppsList />
                </div>
            }
        />
    );
}

/** The apps given the node's Docker socket, each linked to the screen that takes it away. */
function PrivilegedAppsList() {
    const { data } = HivePaaSSecuritySettingsQueries.useFindOne();
    const apps = data?.data.privilegedApps ?? [];

    if (apps.length === 0) {
        return <p className="text-muted-foreground">No app has the node&apos;s Docker socket.</p>;
    }
    return (
        <div className="flex flex-col gap-1.5">
            <p>Apps with the node&apos;s Docker socket:</p>
            <ul className="flex flex-col gap-1">
                {apps.map(app => (
                    <li
                        key={app.appId}
                        className="flex items-center gap-2"
                    >
                        <ServerCog className="size-4 shrink-0 text-destructive" />
                        <AppLink.Modules
                            to={ROUTE.projects.single.apps.single.configuration.dockerApi.$route(
                                app.projectId,
                                app.projectEnvKey,
                                app.appId,
                            )}
                            className="text-link"
                        >
                            {app.appName}
                        </AppLink.Modules>
                        <span className="text-muted-foreground">
                            {app.projectName} / {app.projectEnvName}
                        </span>
                    </li>
                ))}
            </ul>
        </div>
    );
}
