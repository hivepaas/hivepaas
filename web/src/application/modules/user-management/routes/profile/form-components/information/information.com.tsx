import { Field, FieldError, FieldGroup, FieldLabel, Input } from "@components/ui";
import { Textarea } from "@components/ui/textarea";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { type ProfileFormSchemaInput, type ProfileFormSchemaOutput } from "../../schemas";

export function Information() {
    const {
        control,
        formState: { errors },
    } = useFormContext<ProfileFormSchemaInput, unknown, ProfileFormSchemaOutput>();

    const {
        field: fullName,
        fieldState: { invalid: isFullNameInvalid },
    } = useController({
        control,
        name: "fullName",
    });

    const {
        field: username,
        fieldState: { invalid: isUsernameInvalid },
    } = useController({
        control,
        name: "username",
    });

    const { field: email } = useController({
        control,
        name: "email",
    });

    const {
        field: position,
        fieldState: { invalid: isPositionInvalid },
    } = useController({
        control,
        name: "position",
    });

    const {
        field: notes,
        fieldState: { invalid: isNotesInvalid },
    } = useController({
        control,
        name: "notes",
        defaultValue: "",
    });

    return (
        <InfoBlock title="Information">
            <FieldGroup>
                <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
                    <Field>
                        <FieldLabel htmlFor="username">Username</FieldLabel>
                        <Input
                            {...username}
                            value={username.value}
                            onChange={username.onChange}
                            type="text"
                            placeholder=""
                            aria-invalid={isUsernameInvalid}
                        />
                        <FieldError errors={[errors.username]} />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="email">Email</FieldLabel>
                        <Input
                            {...email}
                            value={email.value}
                            onChange={email.onChange}
                            type="email"
                            placeholder="tiendc@gmail.com"
                            disabled
                        />
                        <FieldError errors={[errors.email]} />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="fullName">Full name</FieldLabel>
                        <Input
                            {...fullName}
                            value={fullName.value}
                            onChange={fullName.onChange}
                            type="text"
                            placeholder="Enter your full name"
                            aria-invalid={isFullNameInvalid}
                        />
                        <FieldError errors={[errors.fullName]} />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="position">Position</FieldLabel>
                        <Input
                            {...position}
                            value={position.value}
                            onChange={position.onChange}
                            type="text"
                            placeholder="Enter your position"
                            aria-invalid={isPositionInvalid}
                        />
                        <FieldError errors={[errors.position]} />
                    </Field>
                </div>

                <Field>
                    <FieldLabel htmlFor="notes">Notes</FieldLabel>
                    <Textarea
                        {...notes}
                        value={notes.value}
                        onChange={notes.onChange}
                        rows={4}
                        placeholder="Enter notes"
                        aria-invalid={isNotesInvalid}
                    />
                    <FieldError errors={[errors.notes]} />
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}
