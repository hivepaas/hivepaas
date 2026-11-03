import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { AppCloneSettingsFormSchemaInput } from "../schemas";

export interface AppCloneSettingsFormRef {
    submit: () => void;
    setValues: (values: Partial<AppCloneSettingsFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
