import { useEffect, useMemo, useState } from "react";

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppTerminalQueries } from "~/projects/data";

import { AppLink, AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { AppTerminalPanel } from "../building-blocks";

export function AppTerminalRoute() {
    const { id: projectID, appId: appID } = useParams<{ id: string; appId: string }>();
    const [selectedShell, setSelectedShell] = useState("");

    invariant(projectID, "projectID must be defined");
    invariant(appID, "appID must be defined");

    const { data: infoResponse, isLoading: isInfoLoading } = AppTerminalQueries.useGetInfo({
        projectID,
        appID,
    });
    const isTerminalEnabled = infoResponse?.data.enabled ?? true;
    const supportedShells = useMemo(
        () => infoResponse?.data.supportedShells ?? [],
        [infoResponse?.data.supportedShells],
    );

    useEffect(() => {
        setSelectedShell(current => {
            if (supportedShells.includes(current)) {
                return current;
            }

            return supportedShells[0] ?? "";
        });
    }, [supportedShells]);

    return (
        <section className={cn(listBox)}>
            {isInfoLoading ? (
                <AppLoader />
            ) : isTerminalEnabled ? (
                <AppTerminalPanel
                    projectID={projectID}
                    appID={appID}
                    supportedShells={supportedShells}
                    selectedShell={selectedShell}
                    onSelectedShellChange={setSelectedShell}
                />
            ) : (
                <p className="text-base">
                    Terminal feature is disabled, enable it in{" "}
                    <AppLink.Basic
                        to={ROUTE.projects.single.apps.single.configuration.featureSettings.$route(projectID, appID)}
                        className="text-primary underline-offset-4 hover:underline"
                    >
                        Feature Settings
                    </AppLink.Basic>
                </p>
            )}
        </section>
    );
}
