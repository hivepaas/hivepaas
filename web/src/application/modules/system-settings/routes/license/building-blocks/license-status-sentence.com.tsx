import type { ReactNode } from "react";

import type { License } from "@application/shared/entities";
import { ELicenseStatus } from "@application/shared/enums";
import { daysLeftText, licensePlanName } from "@application/shared/utils";

import { LicenseDate } from "./license-date.com";
import { isOnBee } from "./license-labels";

/** "until Oct 31, 2026 (3 days left)", or nothing when the date is not known. */
function until(date: Date | null): ReactNode {
    if (!date) {
        return null;
    }

    return (
        <>
            {" until "}
            <LicenseDate date={date} /> ({daysLeftText(date)})
        </>
    );
}

/** " on Oct 31, 2026", or nothing when the date is not known. */
function on(date: Date | null): ReactNode {
    if (!date) {
        return null;
    }

    return (
        <>
            {" on "}
            <LicenseDate date={date} />
        </>
    );
}

/**
 * What the status means for this installation, in a sentence. The banners say
 * the same in fewer words.
 */
function statusSentence(license: License): ReactNode {
    const { expiresAt, graceEndsAt, movedAt } = license;
    const plan = licensePlanName(license.plan);

    switch (license.status) {
        case ELicenseStatus.None:
            return "This installation runs on Bee, the free plan.";
        case ELicenseStatus.Valid:
            return (
                <>
                    {expiresAt && (
                        <>
                            Paid until <LicenseDate date={expiresAt} />.{" "}
                        </>
                    )}
                    With online renewal on, the renewed key arrives by itself; otherwise paste it below.
                </>
            );
        case ELicenseStatus.Trial:
            return (
                <>
                    A trial of {licensePlanName(license.keyPlan || license.plan)}
                    {until(expiresAt)}. Then this installation is on Bee.
                </>
            );
        case ELicenseStatus.Grace:
            return (
                <>
                    The paid period ended{on(expiresAt)}. {plan} holds{until(graceEndsAt)}, then this installation is on
                    Bee.
                </>
            );
        case ELicenseStatus.Expired:
            return (
                <>
                    The key ran out{on(graceEndsAt ?? expiresAt)}: this installation is on Bee. Nothing that runs was
                    stopped; nothing more can be added past Bee&apos;s limits.
                </>
            );
        case ELicenseStatus.Invalid:
            return "The stored key could not be verified: this installation is on Bee.";
        case ELicenseStatus.OtherInstallation:
            return (
                "The stored key is for another installation: this one is on Bee. A swarm made again has a new ID: " +
                "move the license to it in the portal, then paste the key it issues."
            );
        case ELicenseStatus.Moved:
            return (
                <>
                    The portal moved this key to another installation{on(movedAt)}.{" "}
                    {isOnBee(license) ? (
                        "This installation is now on Bee."
                    ) : (
                        <>It holds here{until(graceEndsAt)}, then this installation is on Bee.</>
                    )}
                </>
            );
        default:
            return null;
    }
}

export function LicenseStatusSentence({ license }: { license: License }) {
    const sentence = statusSentence(license);

    if (!sentence) {
        return null;
    }

    return <p className="text-sm leading-relaxed text-muted-foreground">{sentence}</p>;
}
