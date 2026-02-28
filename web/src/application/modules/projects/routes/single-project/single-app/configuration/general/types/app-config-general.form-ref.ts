import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigGeneralFormSchemaInput } from "../schemas";

export interface AppConfigGeneralFormRef {
    setValues: (values: Partial<AppConfigGeneralFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
