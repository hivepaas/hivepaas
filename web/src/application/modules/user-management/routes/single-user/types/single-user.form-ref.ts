import { type SingleUserFormSchemaInput } from "~/user-management/routes/single-user/schemas";

import { type ValidationException } from "@infrastructure/exceptions/validation";

export interface SingleUserFormRef {
    setValues: (values: SingleUserFormSchemaInput) => void;
    onError: (error: ValidationException) => void;
}
