import { Checkbox, FieldError, Input } from "@components/ui";
import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { useController, useFormContext } from "react-hook-form";
import { EAppArmorMode, ESeccompMode } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    type AppConfigContainerSettingsFormSchemaInput,
    type AppConfigContainerSettingsFormSchemaOutput,
} from "../schemas";

export function SecurityFields() {
    const { control, watch } = useFormContext<
        AppConfigContainerSettingsFormSchemaInput,
        unknown,
        AppConfigContainerSettingsFormSchemaOutput
    >();

    const { field: noNewPrivileges } = useController({ control, name: "privileges.noNewPrivileges" });
    const { field: selinuxEnabled } = useController({ control, name: "privileges.selinuxEnabled" });
    const {
        field: seLinuxUser,
        fieldState: { error: seLinuxUserError },
    } = useController({
        control,
        name: "privileges.seLinuxUser",
    });
    const {
        field: seLinuxRole,
        fieldState: { error: seLinuxRoleError },
    } = useController({
        control,
        name: "privileges.seLinuxRole",
    });
    const {
        field: seLinuxType,
        fieldState: { error: seLinuxTypeError },
    } = useController({
        control,
        name: "privileges.seLinuxType",
    });
    const {
        field: seLinuxLevel,
        fieldState: { error: seLinuxLevelError },
    } = useController({
        control,
        name: "privileges.seLinuxLevel",
    });
    const { field: seccompMode } = useController({ control, name: "privileges.seccompMode" });
    const {
        field: seccompProfile,
        fieldState: { error: seccompProfileError },
    } = useController({
        control,
        name: "privileges.seccompProfile",
    });
    const { field: appArmorMode } = useController({ control, name: "privileges.appArmorMode" });

    const seccompModeValue = watch("privileges.seccompMode");

    return (
        <>
            <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Security</h3>
            <div className="flex flex-col gap-6 px-2">
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="No New Privileges"
                            content="Sets no-new-privileges on the container."
                        />
                    }
                >
                    <div className="flex items-center gap-2">
                        <Checkbox
                            checked={noNewPrivileges.value}
                            onCheckedChange={v => {
                                noNewPrivileges.onChange(v === true);
                            }}
                        />
                    </div>
                </InfoBlock>

                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="SE Linux Context"
                            content="Enable or disable SELinux labeling for the container."
                        />
                    }
                >
                    <Tabs
                        value={selinuxEnabled.value ? "enabled" : "disabled"}
                        onValueChange={v => {
                            selinuxEnabled.onChange(v === "enabled");
                        }}
                        className="w-fit"
                    >
                        <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                            <TabsTrigger value="enabled">Enabled</TabsTrigger>
                            <TabsTrigger value="disabled">Disabled</TabsTrigger>
                        </TabsList>
                    </Tabs>
                </InfoBlock>

                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="SE Linux Context Settings"
                            content="User, role, type, and level for SELinux."
                        />
                    }
                >
                    <div className="flex flex-wrap gap-3 max-w-[640px]">
                        <Input
                            {...seLinuxUser}
                            value={seLinuxUser.value}
                            onChange={seLinuxUser.onChange}
                            placeholder="user"
                            className="max-w-[140px]"
                            disabled={!selinuxEnabled.value}
                        />
                        <Input
                            {...seLinuxRole}
                            value={seLinuxRole.value}
                            onChange={seLinuxRole.onChange}
                            placeholder="role"
                            className="max-w-[140px]"
                            disabled={!selinuxEnabled.value}
                        />
                        <Input
                            {...seLinuxType}
                            value={seLinuxType.value}
                            onChange={seLinuxType.onChange}
                            placeholder="type"
                            className="max-w-[140px]"
                            disabled={!selinuxEnabled.value}
                        />
                        <Input
                            {...seLinuxLevel}
                            value={seLinuxLevel.value}
                            onChange={seLinuxLevel.onChange}
                            placeholder="level"
                            className="max-w-[140px]"
                            disabled={!selinuxEnabled.value}
                        />
                    </div>
                    <FieldError errors={[seLinuxUserError, seLinuxRoleError, seLinuxTypeError, seLinuxLevelError]} />
                </InfoBlock>

                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="Seccomp Mode"
                            content="Restrict or allow system calls."
                        />
                    }
                >
                    <Tabs
                        value={seccompMode.value}
                        onValueChange={v => {
                            seccompMode.onChange(v as ESeccompMode);
                        }}
                        className="w-fit"
                    >
                        <TabsList className="bg-zinc-100/80 p-1 rounded-lg flex-wrap h-auto">
                            <TabsTrigger value={ESeccompMode.Default}>Default</TabsTrigger>
                            <TabsTrigger value={ESeccompMode.Unconfined}>Unconfined</TabsTrigger>
                            <TabsTrigger value={ESeccompMode.Custom}>Custom</TabsTrigger>
                        </TabsList>
                    </Tabs>
                </InfoBlock>

                {seccompModeValue === ESeccompMode.Custom && (
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Seccomp Profile"
                                content="JSON profile path or inline profile when mode is Custom."
                            />
                        }
                    >
                        <Input
                            {...seccompProfile}
                            value={seccompProfile.value}
                            onChange={seccompProfile.onChange}
                            placeholder="profile"
                            className="max-w-[400px]"
                        />
                        <FieldError errors={[seccompProfileError]} />
                    </InfoBlock>
                )}

                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="AppArmor Mode"
                            content="AppArmor confinement for the container."
                        />
                    }
                >
                    <Tabs
                        value={appArmorMode.value}
                        onValueChange={v => {
                            appArmorMode.onChange(v as EAppArmorMode);
                        }}
                        className="w-fit"
                    >
                        <TabsList className="bg-zinc-100/80 p-1 rounded-lg">
                            <TabsTrigger value={EAppArmorMode.Default}>Default</TabsTrigger>
                            <TabsTrigger value={EAppArmorMode.Disabled}>Disabled</TabsTrigger>
                        </TabsList>
                    </Tabs>
                </InfoBlock>
            </div>
        </>
    );
}
