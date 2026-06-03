import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { ProjectDomainSettingsFormSchemaInput } from "../schemas";

export type ProjectDomainSettingsFormRef = {
    setValues: (values: Partial<ProjectDomainSettingsFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
};
