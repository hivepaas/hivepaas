import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { HivePaaSGeneralFormInput } from "../schemas";

export type HivePaaSGeneralFormRef = {
    setValues: (values: Partial<HivePaaSGeneralFormInput>) => void;
    onError: (error: ValidationException) => void;
};
