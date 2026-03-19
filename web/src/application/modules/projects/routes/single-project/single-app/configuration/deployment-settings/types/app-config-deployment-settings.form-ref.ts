import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigDeploymentSettingsFormSchemaInput } from "../schemas";

export interface AppConfigDeploymentSettingsFormRef {
    setValues: (values: Partial<AppConfigDeploymentSettingsFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
