import type { AppTemplateDetail, AppTemplateDockerApi } from "../api";

import { componentLabel } from "./components";

/** One app of a deploy request and the Docker API deploying it gives that app. */
export interface GrantedDockerApi {
    app: string;
    access: AppTemplateDockerApi;
}

/** HivePaaS's limits for an app whose template names none. */
const DEFAULT_LIMITS = { containers: 5, memory: "1gb", cpus: 1 };

/** What a group of endpoints lets the app's children do, in a few words. */
const GROUP_WORDS: Record<string, string> = {
    exec: "run commands in them",
    files: "copy files in and out",
    volumes: "create volumes",
    networks: "create networks",
    nestedSocket: "give them the Docker API too",
};

/**
 * What a Docker API block lets the app do, in words: one short phrase per part,
 * so the decision is made on what is listed rather than on the words "Docker
 * API".
 */
export function describeDockerApi(access: AppTemplateDockerApi): string[] {
    const images = access.images.includes("*") ? "any image" : access.images.join(", ");
    const described = [`runs ${images}`];
    for (const dir of access.sharedDirs ?? []) {
        described.push(`shares ${dir}`);
    }
    for (const [name, dir] of Object.entries(access.sharedVolumes ?? {})) {
        described.push(`shares ${dir} as the volume ${name}`);
    }
    if (access.networks?.includes("env")) {
        described.push("joins the env network");
    }
    for (const group of access.allow ?? []) {
        described.push(GROUP_WORDS[group] ?? group);
    }
    // The server leaves a limit out where the template keeps the default.
    const containers = access.containers ?? DEFAULT_LIMITS.containers;
    const memory = access.memory ?? DEFAULT_LIMITS.memory;
    const cpus = access.cpus ?? DEFAULT_LIMITS.cpus;
    described.push(`up to ${containers} containers of ${memory}, ${cpus} CPU${cpus === 1 ? "" : "s"}`);
    return described;
}

/**
 * Everything deploying a template would give through the Docker API: its own
 * block, or each component's, and each dependency's, since those apps are
 * created by the same request behind the same permission.
 */
export function grantedDockerApiByTemplate(template: AppTemplateDetail): GrantedDockerApi[] {
    const granted: GrantedDockerApi[] = [];
    if (template.dockerApi) {
        granted.push({ app: template.title, access: template.dockerApi });
    }
    for (const component of template.components ?? []) {
        if (component.dockerApi) {
            granted.push({ app: componentLabel(template, component), access: component.dockerApi });
        }
    }
    for (const dep of template.dependencies ?? []) {
        if (dep.dockerApi) {
            granted.push({ app: dep.templateTitle ?? dep.title, access: dep.dockerApi });
        }
    }
    return granted;
}
