import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { SystemRegistryAuthRenewalConfigurationFormInput } from "../schemas";

export interface SystemRegistryAuthRenewalConfigurationFormRef {
    setValues: (values: Partial<SystemRegistryAuthRenewalConfigurationFormInput>) => void;
    onError: (error: ValidationException) => void;
}
