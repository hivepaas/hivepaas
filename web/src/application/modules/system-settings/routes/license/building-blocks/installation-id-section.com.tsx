import { useId } from "react";

import { CopyField, SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock } from "@application/shared/components";

interface Props {
    /** The swarm's cluster ID. Empty when Docker could not say. */
    installationId: string;
}

/** The ID a key is bought for: checkout and the portal ask for it. */
export function InstallationIdSection({ installationId }: Props) {
    const headingId = useId();

    return (
        <section
            aria-labelledby={headingId}
            className="flex flex-col gap-4"
        >
            <SectionHeader>
                <h2 id={headingId}>Installation</h2>
            </SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <InfoBlock
                    titleWidth={220}
                    title="Installation ID"
                >
                    <div className="flex max-w-[640px] flex-col gap-1.5">
                        {installationId ? (
                            <CopyField
                                what="Installation ID"
                                value={installationId}
                            />
                        ) : (
                            <span className="text-sm text-amber-700 dark:text-amber-400">
                                Docker did not say this swarm&apos;s ID
                            </span>
                        )}
                        <p className="text-xs text-muted-foreground">
                            A key names the installation it is for: checkout and the portal ask for this ID.
                        </p>
                    </div>
                </InfoBlock>
            </div>
        </section>
    );
}
