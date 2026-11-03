import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { SettingsAppPlacementFormSchemaInput } from "../schemas";

export interface SettingsAppPlacementFormRef {
    setValues: (values: Partial<SettingsAppPlacementFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
