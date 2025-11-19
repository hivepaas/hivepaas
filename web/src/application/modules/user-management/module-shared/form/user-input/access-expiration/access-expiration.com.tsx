import React from "react";

import { DateTimePicker } from "@components/ui/date-time-picker";
import { type Path, useController, useFormContext } from "react-hook-form";

function View<T>({ name, className }: Props<T>) {
    const { control } = useFormContext<Record<string, Date | null>>();

    const {
        field: accessExpireAt,
        fieldState: { invalid },
    } = useController({
        control,
        name: name as string,
    });

    return (
        <DateTimePicker
            value={accessExpireAt.value ?? undefined}
            onChange={date => {
                accessExpireAt.onChange(date ?? null);
            }}
            className={className}
            displayFormat={{ hour24: "yyyy-MM-dd HH:mm:ss" }}
            aria-invalid={invalid}
        />
    );
}

interface Props<T> {
    name: Path<T>;
    className?: string;
}

export const AccessExpiration = React.memo(View) as typeof View;
