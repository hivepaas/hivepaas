import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { StorageMountFormInput } from "../schemas";

export interface StorageMountFormRef {
    setValues: (values: Partial<StorageMountFormInput>) => void;
    onError: (error: ValidationException) => void;
}
