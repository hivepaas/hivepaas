import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

/**
 * What TLS passthrough leaves to the app: HivePaaS forwards the encrypted
 * traffic and mounts nothing, so the certificate is mounted by hand.
 */
export function TlsPassthroughMountNote({ className }: Props) {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const settingMounts =
        projectId && env && appId
            ? ROUTE.projects.single.apps.single.configuration.settingMounts.$route(projectId, env, appId)
            : null;

    return (
        <div className={cn(dashedBorderBox, "text-sm", className)}>
            <span className="font-semibold text-orange-500">Note:</span> TLS passthrough forwards encrypted traffic
            untouched, so the app must terminate TLS itself. Mount the domain&apos;s certificate and private key into
            the container in{" "}
            {settingMounts ? (
                <AppLink.Modules
                    to={settingMounts}
                    className="font-semibold text-link hover:underline"
                >
                    Setting Mounts
                </AppLink.Modules>
            ) : (
                <strong>Setting Mounts</strong>
            )}
            , then point the app&apos;s TLS configuration at those files. Mounting the private key takes the Can Reveal
            Secrets permission.
        </div>
    );
}

interface Props {
    className?: string;
}
