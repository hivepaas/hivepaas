/**
 * What the stored license key amounts to on this installation, now. Only valid,
 * trial and grace, and moved until its grace runs out, put the key's plan in
 * force; every other status leaves the installation on Bee.
 */
export const ELicenseStatus = {
    /** No key stored. */
    None: "none",
    /** A key within its paid period. */
    Valid: "valid",
    /** A trial key within its period. */
    Trial: "trial",
    /** A key past its period and within its grace days. */
    Grace: "grace",
    /** A key past its grace days. */
    Expired: "expired",
    /** A key whose signatures or content do not verify. */
    Invalid: "invalid",
    /** A valid key issued for other installations. */
    OtherInstallation: "other-installation",
    /**
     * A key the portal has since moved to another installation. It holds until its
     * grace runs out, counted from the move, and then the plan turns to Bee while
     * the status stays moved: it never becomes expired.
     */
    Moved: "moved",
} as const;

export type ELicenseStatus = (typeof ELicenseStatus)[keyof typeof ELicenseStatus];
