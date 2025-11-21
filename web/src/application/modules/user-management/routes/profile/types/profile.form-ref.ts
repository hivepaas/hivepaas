import { type ProfileFormSchemaInput } from "~/user-management/routes/profile/schemas";

import { type ValidationException } from "@infrastructure/exceptions/validation";

export interface ProfileFormRef {
    setValues: (values: ProfileFormSchemaInput) => void;
    onError: (error: ValidationException) => void;
}
