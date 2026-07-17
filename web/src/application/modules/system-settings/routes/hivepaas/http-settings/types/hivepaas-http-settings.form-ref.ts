import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { HivePaaSHttpSettingsFormInput } from "../schemas";

export type HivePaaSHttpSettingsFormRef = {
    setValues: (values: Partial<HivePaaSHttpSettingsFormInput>) => void;
    onError: (error: ValidationException) => void;
};
