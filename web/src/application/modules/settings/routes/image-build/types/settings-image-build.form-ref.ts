import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { SettingsImageBuildFormSchemaInput } from "../schemas";

export interface SettingsImageBuildFormRef {
    setValues: (values: Partial<SettingsImageBuildFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
