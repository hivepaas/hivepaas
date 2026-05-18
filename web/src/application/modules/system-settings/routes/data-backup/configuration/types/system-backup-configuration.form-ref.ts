import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { SystemBackupConfigurationFormInput } from "../schemas";

export interface SystemBackupConfigurationFormRef {
    setValues: (values: Partial<SystemBackupConfigurationFormInput>) => void;
    onError: (error: ValidationException) => void;
}
