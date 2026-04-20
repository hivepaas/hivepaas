import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigHttpSettingsFormSchemaInput } from "../schemas";

export interface AppConfigHttpSettingsFormRef {
    setValues: (values: Partial<AppConfigHttpSettingsFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
