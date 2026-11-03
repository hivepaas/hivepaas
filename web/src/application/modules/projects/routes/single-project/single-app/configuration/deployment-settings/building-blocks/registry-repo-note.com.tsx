import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { registryCreatesRepositories } from "./registry-repo-note.utils";

interface Props {
    /** The full reference a build will push, empty when no registry is selected. */
    reference: string;
    /** The address of the selected registry, which decides whether this shows. */
    address: string;
}

export function RegistryRepoNote({ reference, address }: Props) {
    if (!reference || !address || registryCreatesRepositories(address)) {
        return null;
    }

    return (
        <div className={cn(dashedBorderBox)}>
            <span className="font-semibold text-orange-500">Create this repository first:</span>{" "}
            <span className="font-mono">{reference}</span>. This registry does not create repositories on push, so a
            build would fail at its last step.
        </div>
    );
}
