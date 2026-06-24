import { PasswordInput } from "@components/ui/input-password";
import type { Control, FieldErrors, FieldPath } from "react-hook-form";
import { useController } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Field, FieldError, Input } from "@/components/ui";
import { Textarea } from "@/components/ui/textarea";

import type { CreateOrEditAcmeDnsProviderFormInput } from "../create-or-edit-acme-dns-provider.form.schema";

type FieldName = FieldPath<CreateOrEditAcmeDnsProviderFormInput>;

export function AcmeDnsProviderField({ control, errors, name, label, isRequired, type = "text" }: Props) {
    const {
        field,
        fieldState: { invalid },
    } = useController({ name, control });
    const error = errors[name];

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label={label}
                    isRequired={isRequired}
                />
            }
        >
            <Field>
                {type === "textarea" && (
                    <Textarea
                        {...field}
                        value={String(field.value)}
                        aria-invalid={invalid}
                        maxLength={1000}
                        minRows={6}
                        maxRows={10}
                    />
                )}
                {type === "password" && (
                    <PasswordInput
                        value={String(field.value)}
                        onChange={field.onChange}
                        onBlur={field.onBlur}
                        name={field.name}
                        ref={field.ref}
                        aria-invalid={invalid}
                    />
                )}
                {type === "text" && (
                    <Input
                        {...field}
                        value={String(field.value)}
                        aria-invalid={invalid}
                    />
                )}
                <FieldError errors={error ? [error] : []} />
            </Field>
        </InfoBlock>
    );
}

interface Props {
    control: Control<CreateOrEditAcmeDnsProviderFormInput>;
    errors: FieldErrors<CreateOrEditAcmeDnsProviderFormInput>;
    name: FieldName;
    label: string;
    isRequired?: boolean;
    type?: "text" | "password" | "textarea";
}
