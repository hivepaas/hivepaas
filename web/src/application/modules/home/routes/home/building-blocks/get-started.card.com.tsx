import { CircleCheck, CircleX, Lightbulb, Loader2, ShieldCheck, X } from "lucide-react";
import { GetStartedCommands, GetStartedQueries } from "~/home/data";
import type { DashboardCert } from "~/home/domain";
import { ProvisionGithubAppDialog, useProvisionGithubAppDialog } from "~/settings/dialogs/provision-github-app";

import { useProfileContext } from "@application/shared/context";
import { SessionQueries } from "@application/shared/data/queries";
import { EUserRole } from "@application/shared/enums";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

/** The installation step the card is shown for. */
const GET_STARTED_STEP = "hivepaas/get-started";

const REOPEN_BROWSER =
    "Certificate installed. Quit and reopen your browser to see this site as secure: " +
    "a browser keeps the connection it opened with the old certificate, even in a new tab.";

function CertStatusIcon({ cert }: { cert: DashboardCert }) {
    switch (cert.status) {
        case "done":
            return <CircleCheck className="mt-0.5 size-4 shrink-0 text-green-600 dark:text-green-500" />;
        case "failed":
            return <CircleX className="mt-0.5 size-4 shrink-0 text-destructive" />;
        case "obtaining":
            return <Loader2 className="mt-0.5 size-4 shrink-0 animate-spin text-muted-foreground" />;
        default:
            return <span className="mt-1 size-3.5 shrink-0 rounded-full border-2 border-muted-foreground/50" />;
    }
}

interface RowProps {
    icon: React.ReactNode;
    title: string;
    tag?: string;
    children: React.ReactNode;
    action?: React.ReactNode;
}

function Row({ icon, title, tag, children, action }: RowProps) {
    return (
        <li className="flex items-start gap-3 px-5 py-3.5 border-b border-border/60 last:border-b-0">
            {icon}
            {/* The button goes under the text on a phone, beside it on anything wider. */}
            <div className="flex min-w-0 grow flex-col gap-2 sm:flex-row sm:items-start sm:gap-3">
                <div className="flex min-w-0 grow flex-col gap-1">
                    <span className="text-sm font-semibold">
                        {title}
                        {tag && <span className="font-normal text-muted-foreground"> · {tag}</span>}
                    </span>
                    <div className="text-[13px] text-muted-foreground break-words">{children}</div>
                </div>
                {action && <div className="shrink-0">{action}</div>}
            </div>
        </li>
    );
}

function DashboardCertRow({ cert }: { cert: DashboardCert }) {
    const { mutate: requestCert, isPending } = GetStartedCommands.useRequestDashboardCert();
    const domain = cert.domain || "the dashboard's domain";

    let detail: React.ReactNode;
    switch (cert.status) {
        case "done":
            detail = REOPEN_BROWSER;
            break;
        case "obtaining":
            detail = `Getting a certificate for ${domain} from Let's Encrypt. This takes a minute or two.`;
            break;
        case "failed":
            detail = (
                <>
                    <span className="text-destructive">The last attempt failed: {cert.error}</span>
                    <br />
                    {`Check that ${domain} points at this server's IP and that port 80 is open, then try again.`}
                </>
            );
            break;
        default:
            detail = `The dashboard still uses a self-signed certificate, which browsers warn about. For one from Let's Encrypt, ${domain} has to point at this server's IP and port 80 has to be open.`;
    }

    return (
        <Row
            icon={<CertStatusIcon cert={cert} />}
            title="Secure the dashboard"
            action={
                cert.status !== "done" && (
                    <Button
                        size="sm"
                        variant="outline"
                        isLoading={isPending}
                        disabled={isPending || cert.status === "obtaining"}
                        onClick={() => {
                            requestCert();
                        }}
                    >
                        <ShieldCheck className="size-4" />
                        {cert.status === "failed" ? "Try again" : "Get the certificate"}
                    </Button>
                )
            }
        >
            {detail}
        </Row>
    );
}

/** A suggestion, not a step: whether one is connected already is not asked. */
function GithubAppRow({ certDone }: { certDone: boolean }) {
    const dialog = useProvisionGithubAppDialog();

    return (
        <Row
            icon={<Lightbulb className="mt-0.5 size-4 shrink-0 text-muted-foreground" />}
            title="Connect a GitHub App"
            tag="Optional"
            action={
                <Button
                    size="sm"
                    variant="outline"
                    onClick={() => {
                        dialog.actions.open({ type: "settings" });
                    }}
                >
                    Connect
                </Button>
            }
        >
            Sign in with GitHub, and create apps from your repositories.
            {!certDone && (
                <>
                    {" "}
                    <span className="text-amber-600 dark:text-amber-400">
                        GitHub sends its webhooks only to a site with a valid certificate: secure the dashboard first.
                    </span>
                </>
            )}
        </Row>
    );
}

/**
 * What a new installation still has to do, for an admin, while the installation
 * step says so: a certificate for the dashboard a browser trusts, with a GitHub
 * App suggested beside it. The first reading that finds the certificate done
 * clears the step on the server; the card keeps saying so until the page is
 * loaded again, and closing it clears the step for every admin.
 */
export function GetStartedCard() {
    const isAdmin = useProfileContext(state => state.profile?.role === EUserRole.Admin);

    // The profile the dashboard loaded at sign-in, not read again for this.
    const { data: profile } = SessionQueries.useGetProfile({ enabled: isAdmin, staleTime: Infinity });
    const shown = isAdmin && profile?.data.nextStep === GET_STARTED_STEP;

    const { data } = GetStartedQueries.useDashboardCert({ enabled: shown });
    const { mutate: dismiss, isPending: isDismissing } = GetStartedCommands.useDismiss();

    const cert = shown ? data?.data : undefined;
    if (!cert) {
        return null;
    }

    return (
        <Card className="gap-0 py-0">
            <CardHeader className="flex flex-row items-center gap-2 border-b px-5 py-4 [.border-b]:pb-4">
                <CardTitle className="text-[15px]">Get started</CardTitle>
                <Button
                    size="icon-sm"
                    variant="ghost"
                    className="ml-auto"
                    aria-label="Close Get started"
                    title="Close for every admin"
                    disabled={isDismissing}
                    onClick={() => {
                        dismiss();
                    }}
                >
                    <X className="size-4" />
                </Button>
            </CardHeader>
            <CardContent className="px-0">
                <ul>
                    <DashboardCertRow cert={cert} />
                    <GithubAppRow certDone={cert.status === "done"} />
                </ul>
            </CardContent>
            <ProvisionGithubAppDialog />
        </Card>
    );
}
