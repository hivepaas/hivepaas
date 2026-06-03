import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { ProjectImageBuildSettingsFormSchemaInput } from "../schemas";

export interface ProjectImageBuildSettingsFormRef {
    setValues: (values: Partial<ProjectImageBuildSettingsFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
