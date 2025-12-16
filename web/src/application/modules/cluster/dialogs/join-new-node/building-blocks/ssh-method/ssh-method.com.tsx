import React from "react";

import { useController, useFormContext } from "react-hook-form";

import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { type JoinNewNodeFormInput, type JoinNewNodeFormOutput } from "../../schemas";

function View() {
    const { control } = useFormContext<JoinNewNodeFormInput, unknown, JoinNewNodeFormOutput>();
    const {
        field: hostField,
        fieldState: { invalid: isHostInvalid, error: hostError },
    } = useController({
        name: "host",
        control,
    });

    const {
        field: portField,
        fieldState: { invalid: isPortInvalid, error: portError },
    } = useController({
        name: "port",
        control,
    });

    const {
        field: userField,
        fieldState: { invalid: isUserInvalid, error: userError },
    } = useController({
        name: "user",
        control,
    });

    const {
        field: sshKeyField,
        fieldState: { invalid: isSshKeyInvalid, error: sshKeyError },
    } = useController({
        name: "sshKey",
        control,
    });

    return (
        <>
            <Field>
                <FieldLabel htmlFor="host">Host</FieldLabel>
                <Input
                    id="host"
                    {...hostField}
                    placeholder="11.22.33.44"
                    aria-invalid={isHostInvalid}
                />
                <FieldError errors={[hostError]} />
            </Field>

            <Field>
                <FieldLabel htmlFor="port">Port</FieldLabel>
                <Input
                    id="port"
                    {...portField}
                    placeholder="22"
                    aria-invalid={isPortInvalid}
                />
                <FieldError errors={[portError]} />
            </Field>

            <Field>
                <FieldLabel htmlFor="user">User</FieldLabel>
                <Input
                    id="user"
                    {...userField}
                    placeholder="root"
                    aria-invalid={isUserInvalid}
                />
                <FieldError errors={[userError]} />
            </Field>

            <Field>
                <FieldLabel htmlFor="sshKeyId">SSH key</FieldLabel>
                <Select
                    value={sshKeyField.value?.id ?? ""}
                    onValueChange={value => {
                        sshKeyField.onChange({
                            id: value,
                            name:
                                [
                                    { id: "1", name: "SSH Key 1" },
                                    { id: "2", name: "SSH Key 2" },
                                ].find(key => key.id === value)?.name ?? "",
                        });
                    }}
                >
                    <SelectTrigger
                        id="sshKeyId"
                        className="w-full"
                        aria-invalid={isSshKeyInvalid}
                    >
                        <SelectValue placeholder="select ssh key" />
                    </SelectTrigger>
                    <SelectContent>
                        {[
                            { id: "1", name: "SSH Key 1" },
                            { id: "2", name: "SSH Key 2" },
                        ].map(key => (
                            <SelectItem
                                key={key.id}
                                value={key.id}
                            >
                                {key.name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <FieldError errors={[sshKeyError]} />
            </Field>
        </>
    );
}

export const SshMethod = React.memo(View);
