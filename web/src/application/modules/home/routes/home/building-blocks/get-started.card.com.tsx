import { useEffect, useRef } from "react";

import { CircleCheck, CircleX, KeyRound, Loader2, ShieldCheck, X } from "lucide-react";
import { toast } from "sonner";
import { GetStartedCommands } from "~/home/data";
import { ProvisionGithubAppDialog, useProvisionGithubAppDialog } from "~/settings/dialogs/provision-github-app";

import { useProfileContext } from "@application/shared/context";
import { SessionQueries } from "@application/shared/data/queries";
import { useF2aSetupDialog } from "@application/shared/dialogs";
import type { SetupChecklist, SetupChecklistItem } from "@application/shared/entities";
import { EUserRole } from "@application/shared/enums";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

/** How often the profile is read again while the certificate is being obtained. */
const OBTAINING_REFRESH_MS = 5_000;

const REOPEN_BROWSER =
    "Certificate installed. Quit and reopen your browser to see this site as secure: " +
    "a browser keeps the connection it opened with the old certificate, even in a new tab.";

function StatusIcon({ item }: { item: SetupChecklistItem }) {
    switch (item.status) {
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
    item: SetupChecklistItem;
    title: string;
    tag?: string;
    children: React.ReactNode;
    action?: React.ReactNode;
}

function Row({ item, title, tag, children, action }: RowProps) {
    return (
        <li className="flex items-start gap-3 px-5 py-3.5 border-b border-border/60 last:border-b-0">
            <StatusIcon item={item} />
            {/* The button goes under the text on a phone, beside it on anything wider. */}
            <div className="flex min-w-0 grow flex-col gap-2 sm:flex-row sm:items-start sm:gap-3">
                <div className="flex min-w-0 grow flex-col gap-1">
                    <span className="text-sm font-semibold">
                        {title}
                        {tag && <span className="font-normal text-muted-foreground"> · {tag}</span>}
                    </span>
                    <div className="text-[13px] text-muted-foreground break-words">{children}</div>
                </div>
                {action && item.status !== "done" && <div className="shrink-0">{action}</div>}
            </div>
        </li>
    );
}

function DashboardCertRow({ item }: { item: SetupChecklistItem }) {
    const { mutate: requestCert, isPending } = GetStartedCommands.useRequestDashboardCert();
    const domain = item.domain || "the dashboard's domain";

    let detail: React.ReactNode;
    switch (item.status) {
        case "done":
            detail = REOPEN_BROWSER;
            break;
        case "obtaining":
            detail = `Getting a certificate for ${domain} from Let's Encrypt. This takes a minute or two.`;
            break;
        case "failed":
            detail = (
                <>
                    <span className="text-destructive">The last attempt failed: {item.error}</span>
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
            item={item}
            title="Secure the dashboard"
            action={
                <Button
                    size="sm"
                    variant="outline"
                    isLoading={isPending}
                    disabled={isPending || item.status === "obtaining"}
                    onClick={() => {
                        requestCert();
                    }}
                >
                    <ShieldCheck className="size-4" />
                    {item.status === "failed" ? "Try again" : "Get the certificate"}
                </Button>
            }
        >
            {detail}
        </Row>
    );
}

function TwoFactorRow({ item }: { item: SetupChecklistItem }) {
    const dialog = useF2aSetupDialog({
        onClose: () => {
            dialog.actions.close();
        },
    });

    return (
        <Row
            item={item}
            title="Turn on two-factor authentication"
            tag="Recommended"
            action={
                <Button
                    size="sm"
                    variant="outline"
                    onClick={dialog.actions.open}
                >
                    <KeyRound className="size-4" />
                    Set up
                </Button>
            }
        >
            {item.status === "done"
                ? "Your account asks for a code from your authenticator app at sign-in."
                : "Ask for a code from an authenticator app at sign-in, on top of your password."}
        </Row>
    );
}

function GithubAppRow({ item, certDone }: { item: SetupChecklistItem; certDone: boolean }) {
    const dialog = useProvisionGithubAppDialog();

    return (
        <Row
            item={item}
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
            {item.status === "done" ? (
                "A GitHub App is connected."
            ) : (
                <>
                    Sign in with GitHub, and create apps from your repositories.
                    {!certDone && (
                        <>
                            {" "}
                            <span className="text-amber-600 dark:text-amber-400">
                                GitHub sends its webhooks only to a site with a valid certificate: secure the dashboard
                                first.
                            </span>
                        </>
                    )}
                </>
            )}
        </Row>
    );
}

/**
 * Tells the admin when the certificate arrives while they watch, since the row
 * saying so goes with the card once everything is done.
 */
function useCertArrivedToast(checklist: SetupChecklist | null) {
    const previous = useRef(checklist?.dashboardCert.status);
    const status = checklist?.dashboardCert.status;

    useEffect(() => {
        if (previous.current === "obtaining" && status === "done") {
            toast.success(REOPEN_BROWSER, { duration: 15_000 });
        }
        previous.current = status;
    }, [status]);
}

/**
 * What a new installation still has to do, for an admin, until it is done or
 * closed. The server works out each item from what exists, and clears the step
 * for every admin when all three are done or one of them closes the card.
 */
export function GetStartedCard() {
    const isAdmin = useProfileContext(state => state.profile?.role === EUserRole.Admin);

    const { data } = SessionQueries.useGetProfile({
        enabled: isAdmin,
        refetchInterval: query =>
            query.state.data?.data.setupChecklist?.dashboardCert.status === "obtaining" ? OBTAINING_REFRESH_MS : false,
    });
    const { mutate: dismiss, isPending: isDismissing } = GetStartedCommands.useDismiss();

    const checklist = isAdmin ? (data?.data.setupChecklist ?? null) : null;
    useCertArrivedToast(checklist);

    if (!checklist) {
        return null;
    }

    const left = [checklist.dashboardCert, checklist.twoFactor, checklist.githubApp].filter(
        item => item.status !== "done",
    ).length;

    return (
        <Card className="gap-0 py-0">
            <CardHeader className="flex flex-row items-center gap-2 border-b px-5 py-4 [.border-b]:pb-4">
                <CardTitle className="text-[15px]">Get started</CardTitle>
                <Badge
                    variant="secondary"
                    className="rounded-full px-2"
                >
                    {left} left
                </Badge>
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
                    <DashboardCertRow item={checklist.dashboardCert} />
                    <TwoFactorRow item={checklist.twoFactor} />
                    <GithubAppRow
                        item={checklist.githubApp}
                        certDone={checklist.dashboardCert.status === "done"}
                    />
                </ul>
            </CardContent>
            <ProvisionGithubAppDialog />
        </Card>
    );
}
