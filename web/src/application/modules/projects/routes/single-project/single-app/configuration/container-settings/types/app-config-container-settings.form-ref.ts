import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigContainerSettingsFormSchemaInput } from "../schemas";

export interface AppConfigContainerSettingsFormRef {
    setValues: (values: Partial<AppConfigContainerSettingsFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
