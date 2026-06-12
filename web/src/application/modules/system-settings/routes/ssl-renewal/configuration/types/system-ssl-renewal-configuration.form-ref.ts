import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { SystemSslRenewalConfigurationFormInput } from "../schemas";

export interface SystemSslRenewalConfigurationFormRef {
    setValues: (values: Partial<SystemSslRenewalConfigurationFormInput>) => void;
    onError: (error: ValidationException) => void;
}
