import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { SystemCleanupConfigurationFormInput } from "../schemas";

export type SystemCleanupConfigurationFormRef = {
    setValues: (values: Partial<SystemCleanupConfigurationFormInput>) => void;
    onError: (error: ValidationException) => void;
};
