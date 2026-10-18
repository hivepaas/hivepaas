import type { AppTemplateSummary } from "../api";

/** A template's engine, as a person reads it: "PostgreSQL" for postgres. One this dashboard does not know is shown as it is. */
export function managedEngineName(engine: string): string {
    return engine === "postgres" ? "PostgreSQL" : engine;
}

/** The features creating a template needs that the license in force does not grant. */
export function missingFeaturesOf(template: Pick<AppTemplateSummary, "missingFeatures">): string[] {
    return template.missingFeatures ?? [];
}

/** Whether this HivePaaS can create a template's app: it reads the template, and the license grants what it needs. */
export function canDeployTemplate(template: Pick<AppTemplateSummary, "compatible" | "missingFeatures">): boolean {
    return template.compatible && missingFeaturesOf(template).length === 0;
}
