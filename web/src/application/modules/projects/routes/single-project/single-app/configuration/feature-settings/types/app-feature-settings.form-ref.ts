import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { AppFeatureSettingsFormSchemaInput } from "../schemas";

export interface AppFeatureSettingsFormRef {
    setValues: (values: Partial<AppFeatureSettingsFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
