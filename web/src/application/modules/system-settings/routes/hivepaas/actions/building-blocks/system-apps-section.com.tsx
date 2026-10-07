import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Link } from "react-router";
import { HivePaaSProjectQueries } from "~/system-settings/data";

import { ROUTE } from "@application/shared/constants";

import { Button } from "@/components/ui";

/**
 * The way to the apps HivePaaS runs itself, which the projects list leaves out:
 * their logs, metrics and instances are on their own pages, as any app's are.
 */
export function SystemAppsSection() {
    const { data, isLoading } = HivePaaSProjectQueries.useFindOne();
    const project = data?.data;

    return (
        <div className="rounded-lg border bg-background p-4">
            <div className="flex flex-col items-start gap-6">
                <p className="text-sm font-medium text-foreground">
                    The apps HivePaaS runs itself: its backend, worker, proxy, agent and the system apps it provisions.
                    Open them to read their logs, metrics and instances.
                </p>
                <div className={cn(dashedBorderBox)}>
                    <span className="font-semibold text-orange-500">Warning:</span> Changing their settings directly can
                    break HivePaaS. Change them only if you know what the change does; their own pages under System are
                    the safe way to configure them.
                </div>
                {project ? (
                    <Button
                        className="min-w-[120px]"
                        asChild
                    >
                        <Link to={ROUTE.projects.single.apps.$route(project.id)}>View System Apps</Link>
                    </Button>
                ) : (
                    <Button
                        type="button"
                        className="min-w-[120px]"
                        disabled
                        isLoading={isLoading}
                    >
                        View System Apps
                    </Button>
                )}
            </div>
        </div>
    );
}
