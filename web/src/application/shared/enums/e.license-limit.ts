/** A kind of thing a plan limits. */
export const ELicenseLimit = {
    Nodes: "nodes",
    Users: "users",
    Apps: "apps",
    Functions: "functions",
    Previews: "previews",
    /**
     * Google, Microsoft and OpenID Connect sign-in. Not a count: Bee allows none to
     * be added (max 0) and every paid plan allows all (max null).
     */
    SignInProviders: "sign-in-providers",
} as const;

export type ELicenseLimit = (typeof ELicenseLimit)[keyof typeof ELicenseLimit];
