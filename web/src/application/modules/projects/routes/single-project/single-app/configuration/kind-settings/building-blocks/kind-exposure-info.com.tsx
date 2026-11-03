import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppRoutingSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";

import { AppLink, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Badge } from "@/components/ui/badge";

/**
 * Where a database or cache can be reached from outside the cluster. It is only
 * shown here: its domains, their certificates and TLS passthrough are set in the
 * routing settings, the one place that changes them.
 */
export function KindExposureInfo() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppRoutingSettingsQueries.useFindOne(
        { projectID: projectId, env, appID: appId },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const routing = data?.data;
    const domains = routing?.exposePublicly ? routing.domains.filter(d => d.enabled && d.domain) : [];
    const routingLink = (
        <AppLink.Modules
            to={ROUTE.projects.single.apps.single.configuration.routingSettings.$route(projectId, env, appId)}
            className="font-semibold text-link hover:underline"
        >
            Routing Settings
        </AppLink.Modules>
    );

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label="Public Access"
                    content="Where clients outside the cluster reach this service. Its domains, certificates and TLS passthrough are set in Routing Settings."
                />
            }
        >
            <div className="flex flex-col gap-1.5 pt-1.5 text-sm">
                {isLoading ? (
                    <span className="text-muted-foreground">Loading…</span>
                ) : domains.length > 0 ? (
                    <>
                        {domains.map(domain => (
                            <div
                                key={domain.domain}
                                className="flex flex-wrap items-center gap-1.5"
                            >
                                <span className="font-mono">
                                    {domain.domain}:{domain.containerPort || routing?.port}
                                </span>
                                {domain.tlsPassthrough && (
                                    <Badge
                                        variant="outline"
                                        className="text-[10px] px-1.5 py-0"
                                    >
                                        TLS passthrough
                                    </Badge>
                                )}
                            </div>
                        ))}
                        <span className="text-muted-foreground">Change it in {routingLink}.</span>
                    </>
                ) : (
                    <span className="text-muted-foreground">
                        Not reachable from outside the cluster. To expose it, add a TCP domain in {routingLink}.
                    </span>
                )}
            </div>
        </InfoBlock>
    );
}
