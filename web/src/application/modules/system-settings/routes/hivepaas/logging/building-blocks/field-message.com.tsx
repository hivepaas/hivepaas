import { type FieldPath, useFormContext } from "react-hook-form";

import type { HivePaaSLoggingSettingsFormInput } from "../schemas";

/** The validation message for one field, or nothing. */
export function FieldMessage({ name }: { name: FieldPath<HivePaaSLoggingSettingsFormInput> }) {
    const { getFieldState, formState } = useFormContext<HivePaaSLoggingSettingsFormInput>();
    const { error } = getFieldState(name, formState);

    return error?.message ? <p className="mt-1 text-sm text-destructive">{error.message}</p> : null;
}
