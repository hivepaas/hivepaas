import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { TraefikGeneralFormInput } from "../schemas";

export type TraefikGeneralFormRef = {
    setValues: (values: Partial<TraefikGeneralFormInput>) => void;
    onError: (error: ValidationException) => void;
};
