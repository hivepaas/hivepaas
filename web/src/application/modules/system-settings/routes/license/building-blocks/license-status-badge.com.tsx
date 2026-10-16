import { Badge, type StatusMeaning, statusClassName } from "@components/ui/badge";

import type { License } from "@application/shared/entities";
import { ELicenseStatus } from "@application/shared/enums";

import { isOnBee } from "./license-labels";

interface StatusLook {
    label: string;
    meaning: StatusMeaning;
}

function statusLook(license: License): StatusLook {
    switch (license.status) {
        case ELicenseStatus.None:
            return { label: "No license", meaning: "inactive" };
        case ELicenseStatus.Valid:
            return { label: "Active", meaning: "success" };
        case ELicenseStatus.Trial:
            return { label: "Trial", meaning: "info" };
        case ELicenseStatus.Grace:
            return { label: "Grace period", meaning: "attention" };
        case ELicenseStatus.Expired:
            return { label: "Expired", meaning: "failure" };
        case ELicenseStatus.Invalid:
            return { label: "Invalid key", meaning: "failure" };
        case ELicenseStatus.OtherInstallation:
            return { label: "Another installation's key", meaning: "failure" };
        case ELicenseStatus.Moved:
            // It holds through the grace the move leaves, and is on Bee after.
            return { label: "Moved", meaning: isOnBee(license) ? "failure" : "attention" };
        default:
            // A status from a newer HivePaaS, as it is.
            return { label: license.status, meaning: "inactive" };
    }
}

/** What the stored key amounts to here, now. */
export function LicenseStatusBadge({ license }: { license: License }) {
    const { label, meaning } = statusLook(license);

    return <Badge className={statusClassName(meaning)}>{label}</Badge>;
}
