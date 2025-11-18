import { useController, useFormContext } from "react-hook-form";

import { DatePicker } from "@application/shared/components";
import { InfoBlock } from "@application/shared/components";

import { type SingleUserFormSchemaInput, type SingleUserFormSchemaOutput } from "../../schemas";

export function AccessExpiration() {
    const { control } = useFormContext<SingleUserFormSchemaInput, unknown, SingleUserFormSchemaOutput>();

    const {
        field: accessExpireAt,
        fieldState: { invalid },
    } = useController({
        control,
        name: "accessExpireAt",
    });

    return (
        <InfoBlock title="Access Expiration">
            <DatePicker
                className="min-w-[400px] w-fit"
                value={accessExpireAt.value}
                onChange={date => {
                    accessExpireAt.onChange(date ?? null);
                }}
                placeholder="Select expiration date"
                aria-invalid={invalid}
            />
        </InfoBlock>
    );
}
