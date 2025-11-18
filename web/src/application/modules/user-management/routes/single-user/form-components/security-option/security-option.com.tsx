import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";
import { ESecuritySettings } from "@application/shared/enums";

import { type SingleUserFormSchemaInput, type SingleUserFormSchemaOutput } from "../../schemas";

const securityOptionMap: Record<ESecuritySettings, string> = {
    [ESecuritySettings.EnforceSSO]: "Enforce SSO",
    [ESecuritySettings.Password2FA]: "Password + 2FA",
    [ESecuritySettings.PasswordOnly]: "Password Only",
};

export function SecurityOption() {
    const { control } = useFormContext<SingleUserFormSchemaInput, unknown, SingleUserFormSchemaOutput>();

    const {
        field: securityOption,
        fieldState: { invalid },
    } = useController({
        control,
        name: "securityOption",
    });

    return (
        <InfoBlock title="Security Option">
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
        </InfoBlock>
    );
}
