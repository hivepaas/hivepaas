import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigRoutingSettingsFormSchemaInput } from "../schemas";

export interface AppConfigRoutingSettingsFormRef {
    setValues: (values: Partial<AppConfigRoutingSettingsFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}

export type AppConfigHttpSettingsFormRef = AppConfigRoutingSettingsFormRef;
