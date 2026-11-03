import * as React from "react";

/**
 * The id of what names the group a control is in - an InfoBlock's title. A
 * control with no name of its own is named by it: a checkbox or a select alone
 * in a block titled "Enabled" is "Enabled" to a screen reader too.
 */
export const GroupLabelContext = React.createContext<string | undefined>(undefined);

interface Naming {
    "id"?: string;
    "placeholder"?: string;
    "aria-label"?: string;
    "aria-labelledby"?: string;
}

/**
 * The aria-labelledby a control is given: its own, or its group's label when
 * nothing names it - no aria-label, no id a <label> may point at, no placeholder.
 * A control whose content is its value - a select - passes its own id, to be
 * named by both: "Environment development".
 */
export function useGroupLabelledBy(props: Naming, selfId?: string): string | undefined {
    const groupLabel = React.useContext(GroupLabelContext);
    if (props["aria-labelledby"] || props["aria-label"] || props.id || props.placeholder || !groupLabel) {
        return props["aria-labelledby"];
    }

    return selfId ? `${groupLabel} ${selfId}` : groupLabel;
}
