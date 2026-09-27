import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { SystemBackupRepoCleanupConfigurationFormInput } from "../schemas";

export interface SystemBackupRepoCleanupConfigurationFormRef {
    setValues: (values: Partial<SystemBackupRepoCleanupConfigurationFormInput>) => void;
    onError: (error: ValidationException) => void;
}
