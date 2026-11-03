import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type AppConfigNetworksFormSchemaInput } from "../schemas";

export interface AppConfigNetworksFormRef {
    setValues: (values: Partial<AppConfigNetworksFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}
