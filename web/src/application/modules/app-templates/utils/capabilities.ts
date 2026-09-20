import type { AppTemplateCapabilities, AppTemplateDetail } from "../api";

/** One app of a deploy request and what deploying it grants that app. */
export interface GrantedCapabilities {
    app: string;
    capabilities: AppTemplateCapabilities;
}

/**
 * What a capabilities block grants, in words - the names docker uses, which are
 * the names the app's resource settings screen shows for the same block.
 */
export function describeCapabilities(capabilities: AppTemplateCapabilities): string[] {
    const described: string[] = [];
    for (const name of capabilities.capabilityAdd ?? []) {
        described.push(name);
    }
    if (capabilities.enableGPU) {
        described.push("the GPU");
    }
    for (const [name, value] of Object.entries(capabilities.sysctls ?? {})) {
        described.push(`${name} = ${value}`);
    }
    for (const ulimit of capabilities.ulimits ?? []) {
        described.push(`ulimit ${ulimit.name} ${ulimit.soft}/${ulimit.hard}`);
    }
    for (const name of capabilities.capabilityDrop ?? []) {
        described.push(`drops ${name}`);
    }
    return described;
}

/**
 * Everything deploying a template would grant: its own block, and each
 * dependency's - those apps are created by the same request, run on the same
 * nodes, and are gated on the same permission.
 */
export function grantedByTemplate(template: AppTemplateDetail): GrantedCapabilities[] {
    const granted: GrantedCapabilities[] = [];
    if (template.capabilities) {
        granted.push({ app: template.title, capabilities: template.capabilities });
    }
    for (const dep of template.dependencies ?? []) {
        if (dep.capabilities) {
            granted.push({ app: dep.templateTitle ?? dep.title, capabilities: dep.capabilities });
        }
    }
    return granted;
}
