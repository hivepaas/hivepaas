import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigStorageFormSchemaInput } from "../schemas";

export interface AppConfigStorageFormRef {
    setValues: (values: Partial<AppConfigStorageFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
