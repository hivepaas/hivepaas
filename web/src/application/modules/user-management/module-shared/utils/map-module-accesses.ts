import type z from "zod";
import { type AccessSchema } from "~/user-management/module-shared/schemas";

import { MODULES } from "@application/shared/constants";

const DEFAULT_ACCESS = {
    read: false,
    write: false,
    delete: false,
} as const;

export function mapModuleAccesses(
    existingModuleAccess?: z.infer<typeof AccessSchema>[],
): z.infer<typeof AccessSchema>[] {
    return MODULES.map(module => {
        const existingModule = existingModuleAccess?.find(m => m.id === module.id);
        return {
            id: module.id,
            name: module.name,
            access: existingModule?.access ?? DEFAULT_ACCESS,
        };
    });
}
