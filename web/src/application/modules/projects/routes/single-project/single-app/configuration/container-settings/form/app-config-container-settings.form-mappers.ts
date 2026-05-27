import { type AppContainerSettings } from "~/projects/domain";
import { EAppArmorMode, EHealthcheckMode, ERestartPolicyCondition, ESeccompMode } from "~/projects/module-shared/enums";

import { type AppConfigContainerSettingsFormSchemaInput } from "../schemas";

export function mapAppContainerSettingsToFormInput(
    data: AppContainerSettings,
): AppConfigContainerSettingsFormSchemaInput {
    const selinux = data.privileges?.seLinuxContext;
    const selinuxEnabled = selinux == null ? true : !selinux.disable;

    return {
        general: {
            image: data.image,
            command: data.command,
            workingDir: data.workingDir,
            hostname: data.hostname,
            user: data.user,
            groups: data.groups.join(" "),
            tty: data.tty,
            openStdin: data.openStdin,
            readOnly: data.readOnly,
            stopSignal: data.stopSignal,
            stopGracePeriod: data.stopGracePeriod ?? "",
        },
        serviceLabels: Object.entries(data.serviceLabels).map(([key, value]) => ({ key, value })),
        containerLabels: Object.entries(data.containerLabels).map(([key, value]) => ({ key, value })),
        healthcheck: {
            enabled: data.healthcheck?.enabled ?? false,
            mode: data.healthcheck?.mode ?? EHealthcheckMode.Inherit,
            command: data.healthcheck?.command ?? "",
            interval: data.healthcheck?.interval ?? "",
            timeout: data.healthcheck?.timeout ?? "",
            startPeriod: data.healthcheck?.startPeriod ?? "",
            startInterval: data.healthcheck?.startInterval ?? "",
            retries: data.healthcheck?.retries,
        },
        restartPolicy: {
            condition: data.restartPolicy?.condition ?? ERestartPolicyCondition.None,
            delay: data.restartPolicy?.delay ?? "",
            window: data.restartPolicy?.window ?? "",
            maxAttempts: data.restartPolicy?.maxAttempts ?? undefined,
        },
        logDriver: {
            driver: data.logDriver?.name ?? "",
            options: Object.entries(data.logDriver?.options ?? {}).map(([key, value]) => ({ key, value })),
        },
        privileges: {
            noNewPrivileges: data.privileges?.noNewPrivileges ?? false,
            selinuxEnabled,
            seLinuxUser: selinux?.user ?? "",
            seLinuxRole: selinux?.role ?? "",
            seLinuxType: selinux?.type ?? "",
            seLinuxLevel: selinux?.level ?? "",
            seccompMode: data.privileges?.seccomp?.mode ?? ESeccompMode.Default,
            seccompProfile: data.privileges?.seccomp?.profile ?? "",
            appArmorMode: data.privileges?.appArmor?.mode ?? EAppArmorMode.Default,
        },
    };
}
