/**
 * Whether pushing to this registry creates the repository.
 *
 * An address that is not recognized is treated as one that does not. Being told
 * to check something that turns out to be true costs a glance; staying quiet
 * costs a whole build, because the push is its last step.
 */
export function registryCreatesRepositories(address: string): boolean {
    const host = address.trim().toLowerCase().split("/")[0] ?? "";

    if (host === "docker.io" || host === "index.docker.io" || host === "ghcr.io" || host === "quay.io") {
        return true;
    }
    return host.endsWith(".azurecr.io");
}
