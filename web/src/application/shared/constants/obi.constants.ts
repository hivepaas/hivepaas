/** Why a node cannot run OBI, by its agent's preflight, as a person reads it. */
export const OBI_PREFLIGHT_REASON_TEXT: Record<string, string> = {
    "kernel-too-old": "its kernel is older than 5.8",
    "no-btf": "its kernel does not describe its types (no /sys/kernel/btf/vmlinux)",
    "container-virt": "it is a container itself (OpenVZ or LXC), whose kernel is its host's",
    "no-tracefs": "it has no tracefs to attach probes through",
    "lockdown": "its kernel is locked down for confidentiality",
    "low-memory": "too little memory is free: twice what its capacity takes",
};

export function obiPreflightReasonText(reason: string): string {
    return OBI_PREFLIGHT_REASON_TEXT[reason] ?? reason;
}
