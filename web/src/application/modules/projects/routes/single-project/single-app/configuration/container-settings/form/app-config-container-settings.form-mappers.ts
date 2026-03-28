import { type AppContainerSettings } from "~/projects/domain";
import { EAppArmorMode, ERestartPolicyCondition, ESeccompMode } from "~/projects/module-shared/enums";

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
        labels: Object.entries(data.labels).map(([key, value]) => ({ key, value })),
        restartPolicy: {
            condition: data.restartPolicy?.condition ?? ERestartPolicyCondition.None,
            delay: data.restartPolicy?.delay ?? "",
            window: data.restartPolicy?.window ?? "",
            maxAttempts: data.restartPolicy?.maxAttempts != null ? String(data.restartPolicy.maxAttempts) : "",
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
