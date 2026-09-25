import React from "react";

import { CheckIcon, MinusIcon } from "lucide-react";

/**
 * Whether a setting reaches the scope below it - a project's apps, an app's
 * previews - as a narrow icon cell. The title says it in words.
 */
function View({ inheritable, target }: Props) {
    const title = inheritable ? `Available in ${target}` : `Not available in ${target}`;

    return (
        <span
            className="inline-flex items-center justify-center"
            title={title}
        >
            {inheritable ? (
                <CheckIcon className="size-4 text-green-600" />
            ) : (
                <MinusIcon className="size-4 text-muted-foreground" />
            )}
            <span className="sr-only">{title}</span>
        </span>
    );
}

interface Props {
    inheritable: boolean;
    /** "apps" or "previews". */
    target: string;
}

export const InheritableCell = React.memo(View);
