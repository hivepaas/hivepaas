import { EAppScheduledJobArgSeparator } from "../../enums";

import type { CommandArgFormValue, CommandArgGroupFormValue } from "./command-arg-groups-section.types";

export function createDefaultCommandArgGroup(index: number): CommandArgGroupFormValue {
    return {
        enabled: true,
        exportEnv: `CMD_ARG_GROUP_${index + 1}`,
        separator: EAppScheduledJobArgSeparator.Equal,
        args: [],
    };
}

export function createDefaultCommandArg(): CommandArgFormValue {
    return {
        use: true,
        name: "",
        value: "",
    };
}
