import React, { useMemo, useState } from "react";

import { Button } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useController, useFormContext } from "react-hook-form";
import { Link } from "react-router";

import { Combobox, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { SshKeysPublicQueries } from "@application/shared/data/queries";

import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

import { type JoinNewNodeFormInput, type JoinNewNodeFormOutput } from "../../schemas";

function View() {
    const { control } = useFormContext<JoinNewNodeFormInput, unknown, JoinNewNodeFormOutput>();
    const [searchQuery, setSearchQuery] = useState("");

    const { data: { data: sshKeys } = DEFAULT_PAGINATED_DATA, isFetching } = SshKeysPublicQueries.useFindManyPaginated({
        search: searchQuery,
    });

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

    const comboboxOptions = useMemo(() => {
        return sshKeys.map(key => ({
            value: {
                id: key.id,
                name: key.name,
            },
            label: key.name,
        }));
    }, [sshKeys]);

    return (
        <>
            <InfoBlock
                titleWidth={150}
                title="Host"
            >
                <FieldGroup>
                    <Field>
                        <Input
                            id="host"
                            {...hostField}
                            placeholder="11.22.33.44"
                            aria-invalid={isHostInvalid}
                        />
                        <FieldError errors={[hostError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={150}
                title="Port"
            >
                <FieldGroup>
                    <Field>
                        <InputNumber
                            id="port"
                            value={portField.value}
                            onValueChange={portField.onChange}
                            placeholder="22"
                            aria-invalid={isPortInvalid}
                        />
                        <FieldError errors={[portError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={150}
                title="User"
            >
                <FieldGroup>
                    <Field>
                        <Input
                            id="user"
                            {...userField}
                            placeholder="root"
                            aria-invalid={isUserInvalid}
                        />
                        <FieldError errors={[userError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={150}
                title="SSH key"
            >
                <FieldGroup>
                    <Field>
                        <Combobox
                            options={comboboxOptions}
                            value={sshKeyField.value?.id ?? null}
                            onChange={(value, option) => {
                                sshKeyField.onChange(option ?? null);
                            }}
                            onSearch={setSearchQuery}
                            placeholder="Select SSH key"
                            searchable
                            closeOnSelect
                            emptyText="No SSH keys available"
                            className="w-full"
                            valueKey="id"
                            aria-invalid={isSshKeyInvalid}
                            loading={isFetching}
                        />
                        <FieldError errors={[sshKeyError]} />
                        <div className="text-xs">
                            <p>
                                Need to add a new SSH key?{" "}
                                <Link
                                    to="#"
                                    className="text-blue-500"
                                >
                                    Click here
                                </Link>
                            </p>
                        </div>
                    </Field>
                </FieldGroup>
            </InfoBlock>
            <Field>
                <div className={cn(dashedBorderBox, "space-y-6")}>
                    <p className="text-sm">
                        <span className="text-orange-600 dark:text-orange-400 font-semibold">WARNING:</span> If the
                        server above is already part of another Swarm cluster, it will be removed from that cluster. Any
                        associated data will also be deleted.
                    </p>
                </div>
            </Field>
            <Field>
                <div className="flex justify-end">
                    <Button type="submit">Join Node</Button>
                </div>
            </Field>
        </>
    );
}

export const SshMethod = React.memo(View);
