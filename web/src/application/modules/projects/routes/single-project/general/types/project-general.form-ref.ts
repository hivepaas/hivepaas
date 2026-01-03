import { type ValidationException } from "@infrastructure/exceptions/validation";

import { type ProjectGeneralFormSchemaInput } from "../schemas";

export interface ProjectGeneralFormRef {
    setValues: (values: Partial<ProjectGeneralFormSchemaInput>) => void;
    onError: (error: ValidationException) => void;
}

