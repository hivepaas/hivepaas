/** How a check with the license portal for a newer key went. */
export const ELicenseRefreshResult = {
    /** A newer key for this installation, installed: a renewal, a plan change, or a paid key replacing a trial. */
    Newer: "newer",
    /** The key was moved to another installation since. */
    Moved: "moved",
    /** No newer key: the one stored is the one in force. */
    Current: "current",
    /** A key the portal did not issue, such as one signed by the development keys. */
    Unknown: "unknown",
    /** No key stored, so nothing to ask with. */
    None: "none",
    /** The daily check is switched off. Asking now still asks. */
    Disabled: "disabled",
    /** The portal could not be reached or refused, or a newer key could not be installed. */
    Failed: "failed",
} as const;

export type ELicenseRefreshResult = (typeof ELicenseRefreshResult)[keyof typeof ELicenseRefreshResult];
