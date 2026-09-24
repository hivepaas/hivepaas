import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigDockerApiFormSchemaInput } from "../schemas";

export interface AppConfigDockerApiFormRef {
    setValues: (values: Partial<AppConfigDockerApiFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
