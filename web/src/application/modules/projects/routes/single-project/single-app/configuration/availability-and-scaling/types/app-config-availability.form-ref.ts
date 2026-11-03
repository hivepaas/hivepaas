import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigAvailabilitySchemaOutput } from "../schemas";

export interface AppConfigAvailabilityFormRef {
    setValues: (values: Partial<AppConfigAvailabilitySchemaOutput>) => void;
    onError: (error: ValidationException) => void;
}
