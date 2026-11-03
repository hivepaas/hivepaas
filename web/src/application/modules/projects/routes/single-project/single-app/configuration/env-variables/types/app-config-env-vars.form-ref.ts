import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigEnvVarsFormSchemaInput } from "../schemas";

export interface AppConfigEnvVarsFormRef {
    setValues: (values: AppConfigEnvVarsFormSchemaInput) => void;
    onError: (error: ValidationException) => void;
}
