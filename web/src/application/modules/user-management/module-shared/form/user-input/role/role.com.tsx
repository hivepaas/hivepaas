import React from "react";

import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { type Path, useController, useFormContext } from "react-hook-form";

import { EUserRole } from "@application/shared/enums";

const roleMap: Record<EUserRole, string> = {
    [EUserRole.Admin]: "Admin",
    [EUserRole.Member]: "Member",
};

function View<T>({ name }: Props<T>) {
    const { control } = useFormContext<Record<string, EUserRole>>();

    const {
        field: role,
        fieldState: { invalid },
    } = useController({
        control,
        name: name as string,
    });

    return (
        <Tabs
            value={role.value}
            onValueChange={value => {
                role.onChange(value as EUserRole);
            }}
        >
            <TabsList>
                {Object.entries(roleMap).map(([value, label]) => (
                    <TabsTrigger
                        key={value}
                        value={value}
                        className="flex-1"
                        aria-invalid={invalid}
                    >
                        {label}
                    </TabsTrigger>
                ))}
            </TabsList>
        </Tabs>
    );
}

interface Props<T> {
    name: Path<T>;
}

export const Role = React.memo(View) as typeof View;
