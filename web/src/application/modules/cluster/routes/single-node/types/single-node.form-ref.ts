import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type SingleNodeFormSchemaInput } from "../schemas";

export interface SingleNodeFormRef {
    setValues: (values: Partial<SingleNodeFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
