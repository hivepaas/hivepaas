/**
 * A plan, from the free one up. Bee is what an installation runs without a key in
 * force, and no key ever names it.
 */
export const ELicensePlan = {
    Bee: "bee",
    Comb: "comb",
    Hive: "hive",
    Colony: "colony",
    Apiary: "apiary",
} as const;

export type ELicensePlan = (typeof ELicensePlan)[keyof typeof ELicensePlan];
