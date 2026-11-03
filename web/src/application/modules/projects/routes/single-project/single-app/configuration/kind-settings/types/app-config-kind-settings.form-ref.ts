import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigKindSettingsFormSchemaInput } from "../schemas";

export interface AppConfigKindSettingsFormRef {
    setValues: (values: Partial<AppConfigKindSettingsFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
