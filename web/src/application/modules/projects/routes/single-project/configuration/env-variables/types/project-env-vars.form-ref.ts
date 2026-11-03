import { type ProjectEnvVarsFormSchemaInput } from "~/projects/routes/single-project/configuration/env-variables/schemas";

import { type ValidationException } from "@infrastructure/exceptions/validation";

export interface ProjectEnvVarsFormRef {
    setValues: (values: ProjectEnvVarsFormSchemaInput) => void;
    onError: (error: ValidationException) => void;
}
