import { SecuritySwitchSection } from "./security-switch-section.com";

export function AllowPrivilegedAppsSection() {
    return (
        <SecuritySwitchSection
            field="allowPrivilegedApps"
            title="Privileged Apps"
            action="Privileged Apps"
            description={
                <div className="flex flex-col gap-2">
                    <p>
                        Privileged apps reach into the server they run on - the Docker socket, or a directory of the
                        node. Some tools cannot work any other way, such as database cluster managers and CI runners.
                    </p>
                    <p className="text-amber-700 dark:text-amber-400">
                        Such an app has full control of its node, and on a manager node of the whole cluster, including
                        HivePaaS itself and the data of every other app. Keep this disabled unless you need one.
                    </p>
                </div>
            }
        />
    );
}
