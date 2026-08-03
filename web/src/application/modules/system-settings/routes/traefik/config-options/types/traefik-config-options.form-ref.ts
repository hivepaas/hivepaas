import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { TraefikConfigOptionsFormInput } from "../schemas";

export type TraefikConfigOptionsFormRef = {
    setValues: (values: Partial<TraefikConfigOptionsFormInput>) => void;
    onError: (error: ValidationException) => void;
};
