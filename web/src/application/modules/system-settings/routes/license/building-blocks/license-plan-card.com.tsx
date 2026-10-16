import { type PropsWithChildren, type ReactNode, useId } from "react";

import { Badge } from "@components/ui/badge";
import { ExternalLink } from "lucide-react";

import { InfoBlock } from "@application/shared/components";
import type { License } from "@application/shared/entities";
import { ELicenseStatus } from "@application/shared/enums";
import { licensePlanName } from "@application/shared/utils";

import { Button } from "@/components/ui";

import { LicenseDate } from "./license-date.com";
import { isOnBee } from "./license-labels";
import { LicenseStatusBadge } from "./license-status-badge.com";
import { LicenseStatusSentence } from "./license-status-sentence.com";

interface DetailRow {
    title: string;
    value: ReactNode;
}

/** The same instant: a key with no grace days ends its grace when it expires. */
function sameInstant(a: Date | null, b: Date | null): boolean {
    return a !== null && b !== null && a.getTime() === b.getTime();
}

/**
 * Which key is stored. Its own fields are there whenever its signatures
 * verified, in force or not, so a key for another installation or past its
 * grace can still be told apart from the one to paste instead.
 */
function keyDetailRows(license: License): DetailRow[] {
    const { status, keyPlan, plan, trial, issuedAt, expiresAt, graceEndsAt, movedAt, reason, installations } = license;
    const rows: DetailRow[] = [];

    if (license.customerName) {
        rows.push({ title: "Licensed to", value: license.customerName });
    }
    if (license.licenseId) {
        rows.push({
            title: "License ID",
            value: <span className="font-mono text-xs break-all">{license.licenseId}</span>,
        });
    }
    // Only when the key is not what is in force: once it ran out, or when it is
    // for another installation.
    if (keyPlan && keyPlan !== plan) {
        rows.push({ title: "Key plan", value: `${licensePlanName(keyPlan)}${trial ? " trial" : ""}` });
    }
    if (issuedAt) {
        rows.push({ title: "Issued", value: <LicenseDate date={issuedAt} /> });
    }
    if (expiresAt) {
        rows.push({ title: trial ? "Trial ends" : "Paid until", value: <LicenseDate date={expiresAt} /> });
    }
    if (graceEndsAt && !sameInstant(graceEndsAt, expiresAt)) {
        rows.push({ title: "Grace ends", value: <LicenseDate date={graceEndsAt} /> });
    }
    if (movedAt) {
        rows.push({ title: "Moved", value: <LicenseDate date={movedAt} /> });
    }
    if (status === ELicenseStatus.Invalid && reason) {
        rows.push({ title: "Reason", value: <span className="whitespace-pre-wrap break-words">{reason}</span> });
    }
    if (status === ELicenseStatus.OtherInstallation && installations.length > 0) {
        rows.push({
            title: "Key is for",
            value: (
                <span className="flex flex-col font-mono text-xs break-all">
                    {installations.map(installation => (
                        <span key={installation}>{installation}</span>
                    ))}
                </span>
            ),
        });
    }

    return rows;
}

function KeyDetails({ license }: { license: License }) {
    const rows = keyDetailRows(license);

    if (rows.length === 0) {
        return null;
    }

    return (
        <div className="flex flex-col gap-1 text-sm">
            {rows.map(row => (
                <InfoBlock
                    key={row.title}
                    titleWidth={220}
                    title={row.title}
                >
                    <span>{row.value}</span>
                </InfoBlock>
            ))}
        </div>
    );
}

/** A page of the license portal, opened in a new tab. */
function PortalLink({ href, primary = false, children }: PropsWithChildren<{ href: string; primary?: boolean }>) {
    return (
        <Button
            asChild
            variant={primary ? "default" : "outline"}
            size="sm"
        >
            <a
                href={href}
                target="_blank"
                rel="noopener noreferrer"
            >
                {children}
                <ExternalLink />
            </a>
        </Button>
    );
}

interface Props {
    license: License;
}

/**
 * The plan in force and what the stored key amounts to: its status, the key's
 * own details, and the portal's pages. Upgrade carries this installation's ID;
 * a trial is offered only while no key was ever installed, as there is one per
 * installation.
 */
export function LicensePlanCard({ license }: Props) {
    const titleId = useId();
    const { portal } = license;

    const upgradeUrl = portal?.pricingUrl;
    const trialUrl = license.status === ELicenseStatus.None ? portal?.trialUrl : undefined;
    const accountUrl = license.licenseId ? portal?.accountUrl : undefined;
    const hasPortalLinks = [upgradeUrl, trialUrl, accountUrl].some(Boolean);

    return (
        <section
            aria-labelledby={titleId}
            className="flex flex-col gap-4 rounded-xl border bg-card px-5 py-4"
        >
            <div className="flex flex-col gap-1.5">
                <h2
                    id={titleId}
                    className="text-[13px] font-medium text-muted-foreground"
                >
                    Plan
                </h2>
                <div className="flex flex-wrap items-center gap-2.5">
                    <span className="text-2xl font-bold">{licensePlanName(license.plan)}</span>
                    <LicenseStatusBadge license={license} />
                    {license.development && (
                        <Badge
                            tone="violet"
                            title="Signed by the development keys: only a HivePaaS in development accepts it"
                        >
                            Development key
                        </Badge>
                    )}
                </div>
                <LicenseStatusSentence license={license} />
            </div>

            <KeyDetails license={license} />

            {hasPortalLinks && (
                <div className="flex flex-wrap items-center gap-2">
                    {upgradeUrl && (
                        <PortalLink
                            href={upgradeUrl}
                            // On Bee a plan is the next step; on a paid one the page is mostly read.
                            primary={isOnBee(license)}
                        >
                            Upgrade
                        </PortalLink>
                    )}
                    {trialUrl && <PortalLink href={trialUrl}>Start a 14-day trial</PortalLink>}
                    {accountUrl && <PortalLink href={accountUrl}>Your licenses</PortalLink>}
                </div>
            )}
        </section>
    );
}
