import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { HivePaaSRoutingSettingsFormInput } from "../schemas";

export type HivePaaSRoutingSettingsFormRef = {
    setValues: (values: Partial<HivePaaSRoutingSettingsFormInput>) => void;
    onError: (error: ValidationException) => void;
};

export type HivePaaSHttpSettingsFormRef = HivePaaSRoutingSettingsFormRef;
