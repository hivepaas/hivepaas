import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigResourcesFormSchemaInput } from "../schemas";

export interface AppConfigResourcesFormRef {
    setValues: (values: Partial<AppConfigResourcesFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
