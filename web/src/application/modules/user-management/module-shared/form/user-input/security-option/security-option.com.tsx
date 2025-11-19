import React from "react";

import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { type Path, useController, useFormContext } from "react-hook-form";

import { ESecuritySettings } from "@application/shared/enums";

const securityOptionMap: Record<ESecuritySettings, string> = {
    [ESecuritySettings.EnforceSSO]: "Enforce SSO",
    [ESecuritySettings.Password2FA]: "Password 2FA",
    [ESecuritySettings.PasswordOnly]: "Password Only",
};

function View<T>({ name }: Props<T>) {
    const { control } = useFormContext<Record<string, ESecuritySettings>>();

    const {
        field: securityOption,
        fieldState: { invalid },
    } = useController({
        control,
        name: name as string,
    });

    return (
        <Tabs
            value={securityOption.value}
            onValueChange={value => {
                securityOption.onChange(value as ESecuritySettings);
            }}
        >
            <TabsList>
                {Object.entries(securityOptionMap).map(([value, label]) => (
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

export const SecurityOption = React.memo(View) as typeof View;
