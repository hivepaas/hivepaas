import { FieldError } from "@components/ui";
import { type FieldPath, useFormContext } from "react-hook-form";

import type { HivePaaSLoggingSettingsFormInput } from "../schemas";

/** The validation message for one field, in the same shape every other form uses. */
export function FieldMessage({ name }: { name: FieldPath<HivePaaSLoggingSettingsFormInput> }) {
    const { getFieldState, formState } = useFormContext<HivePaaSLoggingSettingsFormInput>();
    const { error } = getFieldState(name, formState);

    if (!error?.message) {
        return null;
    }

    return <FieldError errors={[error]} />;
}
