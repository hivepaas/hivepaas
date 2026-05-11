import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsBasicAuthTable, SettingsRegistryAuthTable } from "~/settings/module-shared/components";

function SettingsDummyPage({ title }: Props) {
    return (
        <div className="bg-background rounded-lg p-6">
            <div className="flex flex-col gap-2">
                <h1 className="text-lg font-semibold text-foreground">{title}</h1>
                <p className="text-sm text-muted-foreground">This page is under construction.</p>
            </div>
        </div>
    );
}

interface Props {
    title: string;
}

export function SettingsBasicAuthRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsBasicAuthTable />
        </div>
    );
}

export function SettingsRegistryAuthRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsRegistryAuthTable />
        </div>
    );
}

export function SettingsSslCertificatesRoute() {
    return <SettingsDummyPage title="SSL certificates" />;
}

export function SettingsEmailAccountsRoute() {
    return <SettingsDummyPage title="Email accounts" />;
}

export function SettingsImPlatformsRoute() {
    return <SettingsDummyPage title="IM platforms" />;
}
