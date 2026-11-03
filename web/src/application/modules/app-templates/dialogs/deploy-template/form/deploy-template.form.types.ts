import type { ManagedFormState } from "./managed-fields.state";

export interface DeployTemplateFormState {
    name: string;
    env: string;
    version: string;
    variant: string;
    imageTag: string;
    params: Record<string, unknown>;
    dependencyParams: Record<string, Record<string, string>>;
    /** A managed template's: where its engine archives, and the defaults changed. Ignored for any other template. */
    managed: ManagedFormState;
}
