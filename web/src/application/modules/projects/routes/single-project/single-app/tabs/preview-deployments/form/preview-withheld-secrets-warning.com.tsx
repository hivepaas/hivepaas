import { inlineLink } from "@/lib/styles";
import { AlertTriangle } from "lucide-react";
import type { AppPreviews_PrepareCreate_Res } from "~/projects/api/services";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

type WithheldSecret = AppPreviews_PrepareCreate_Res["data"]["withheldSecrets"][number];

/**
 * Warns of the app's secrets a preview goes without: those not marked
 * Inheritable. The variables built from them are empty in the preview, which
 * shows up as a preview that fails in ways the app does not - so it is said
 * before the preview is made, while turning the flag on is still a click away.
 * Shown whichever way the preview is made: a comment's preview goes without them
 * all the same.
 */
export function PreviewWithheldSecretsWarning({ projectId, env, appId, secrets }: Props) {
    if (secrets.length === 0) {
        return null;
    }

    return (
        <div
            role="status"
            className="flex items-start gap-2.5 rounded-lg border border-amber-500/30 bg-amber-500/[0.07] px-3.5 py-2.5 dark:border-amber-400/25 dark:bg-amber-400/[0.07]"
        >
            <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-500" />
            <div className="space-y-1.5 text-sm text-foreground">
                <p>
                    <span className="font-medium">The preview will not get these secrets.</span>{" "}
                    <span className="text-muted-foreground">
                        They are not inheritable, so the variables using them are empty in it:
                    </span>
                </p>
                <ul className="list-disc space-y-0.5 pl-5">
                    {secrets.map(secret => (
                        <li key={secret.name}>
                            <code>{secret.name}</code>
                            {secret.envVars.length > 0 && (
                                <span className="text-muted-foreground">
                                    , used by{" "}
                                    {secret.envVars.map((name, i) => (
                                        <span key={name}>
                                            {i > 0 && ", "}
                                            <code className="text-foreground">{name}</code>
                                        </span>
                                    ))}
                                </span>
                            )}
                        </li>
                    ))}
                </ul>
                <p className="text-muted-foreground">
                    To give the preview a secret, turn on <span className="text-foreground">Inheritable</span> for it in
                    the app&apos;s{" "}
                    <AppLink.Modules
                        to={ROUTE.projects.single.apps.single.configuration.secrets.$route(projectId, env, appId)}
                        className={inlineLink}
                    >
                        Secrets
                    </AppLink.Modules>
                    .
                </p>
            </div>
        </div>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    secrets: WithheldSecret[];
}
