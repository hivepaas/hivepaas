import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { LocalPaaSGeneralFormInput } from "../schemas";

export type LocalPaaSGeneralFormRef = {
    setValues: (values: Partial<LocalPaaSGeneralFormInput>) => void;
    onError: (error: ValidationException) => void;
};
