import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Checkbox } from "@/components/ui/checkbox";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

import {
    CreateProfileApiKeyFormSchema,
    type CreateProfileApiKeyFormSchemaInput,
    type CreateProfileApiKeyFormSchemaOutput,
} from "../schemas";

export function CreateProfileApiKeyForm({ onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<CreateProfileApiKeyFormSchemaInput, unknown, CreateProfileApiKeyFormSchemaOutput>({
        defaultValues: {
            name: "",
            accessAction: { read: true, write: false, delete: false },
            expireAt: undefined,
        },
        resolver: zodResolver(CreateProfileApiKeyFormSchema),
        mode: "onSubmit",
    });

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({
        name: "name",
        control,
    });

    const { field: accessAction } = useController({
        name: "accessAction",
        control,
    });

    const {
        field: expireAt,
        fieldState: { invalid: isExpireAtInvalid },
    } = useController({
        name: "expireAt",
        control,
    });

    function onValid(values: CreateProfileApiKeyFormSchemaOutput) {
        void onSubmit(values);
    }

    function onInvalid(fieldErrors: FieldErrors<CreateProfileApiKeyFormSchemaOutput>) {
        console.error(fieldErrors);
    }

    return (
        <div className="flex flex-col gap-4">
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
            >
                <FieldGroup>
                    <InfoBlock
                        title={<LabelWithInfo label="Name" />}
                        titleWidth={120}
                    >
                        <Input
                            id="name"
                            value={name.value}
                            onChange={e => {
                                name.onChange(e.target.value);
                            }}
                            aria-invalid={isNameInvalid}
                            placeholder="Enter API key name"
                        />
                        <FieldError errors={[errors.name]} />
                    </InfoBlock>

                    <InfoBlock
                        title={<LabelWithInfo label="Access limit" />}
                        titleWidth={120}
                    >
                        <div className="flex gap-3 mt-2">
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    id="read"
                                    checked={accessAction.value.read}
                                    onCheckedChange={checked => {
                                        accessAction.onChange({
                                            ...accessAction.value,
                                            read: checked === true,
                                        });
                                    }}
                                />
                                <label
                                    htmlFor="read"
                                    className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                >
                                    Read
                                </label>
                            </div>
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    id="write"
                                    checked={accessAction.value.write}
                                    onCheckedChange={checked => {
                                        accessAction.onChange({
                                            ...accessAction.value,
                                            write: checked === true,
                                        });
                                    }}
                                />
                                <label
                                    htmlFor="write"
                                    className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                >
                                    Write
                                </label>
                            </div>
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    id="delete"
                                    checked={accessAction.value.delete}
                                    onCheckedChange={checked => {
                                        accessAction.onChange({
                                            ...accessAction.value,
                                            delete: checked === true,
                                        });
                                    }}
                                />
                                <label
                                    htmlFor="delete"
                                    className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                                >
                                    Delete
                                </label>
                            </div>
                        </div>
                        <FieldError errors={[errors.accessAction]} />
                    </InfoBlock>

                    <InfoBlock
                        title={<LabelWithInfo label="Access expiration" />}
                        titleWidth={120}
                    >
                        <DateTimePicker
                            value={expireAt.value ?? undefined}
                            onChange={date => {
                                expireAt.onChange(date ?? undefined);
                            }}
                            displayFormat={{ hour24: "yyyy-MM-dd HH:mm:ss" }}
                            aria-invalid={isExpireAtInvalid}
                            placeholder="Select expiration date"
                            fromDate={new Date()}
                            showClearButton
                        />
                        <FieldError errors={[errors.expireAt]} />
                    </InfoBlock>
                </FieldGroup>
            </form>
        </div>
    );
}

interface Props {
    onSubmit: (values: CreateProfileApiKeyFormSchemaOutput) => Promise<void> | void;
}
