import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { ProjectStorageSettingsFormSchemaInput } from "../schemas";

export type ProjectStorageSettingsFormRef = {
    setValues: (values: Partial<ProjectStorageSettingsFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
};
