export const QK = {
    "system-settings.hivepaas.service-settings.find-one": "system-settings.hivepaas.service-settings.find-one",
    "system-settings.hivepaas.service-settings.probe": "system-settings.hivepaas.service-settings.probe",
    "system-settings.hivepaas.request-info.find-one": "system-settings.hivepaas.request-info.find-one",
    "system-settings.hivepaas.routing-settings.find-one": "system-settings.hivepaas.routing-settings.find-one",
    // Separate from find-one on purpose: the probe polls on a timer while the
    // confirm dialog is open, and must not fight the page's own cache entry or
    // inherit its retry and notification behaviour.
    "system-settings.hivepaas.routing-settings.probe": "system-settings.hivepaas.routing-settings.probe",
    "system-settings.hivepaas.http-settings.find-one": "system-settings.hivepaas.routing-settings.find-one",
    "system-settings.hivepaas.security-settings.find-one": "system-settings.hivepaas.security-settings.find-one",

    "system-settings.traefik.service-settings.find-one": "system-settings.traefik.service-settings.find-one",
    "system-settings.traefik.config-options.find-one": "system-settings.traefik.config-options.find-one",
    // Separate from find-one for the same reason as the routing probe above.
    "system-settings.traefik.config-options.probe": "system-settings.traefik.config-options.probe",
    "system-settings.backup.find-one": "system-settings.backup.find-one",
    "system-settings.backup-files.find-many-paginated": "system-settings.backup-files.find-many-paginated",
    "system-settings.backup-files.find-one-by-id": "system-settings.backup-files.find-one-by-id",
    "system-settings.cleanup.find-one": "system-settings.cleanup.find-one",
    "system-settings.cleanup.repo-cache.find-one": "system-settings.cleanup.repo-cache.find-one",
    "system-settings.ssl-renewal.find-one": "system-settings.ssl-renewal.find-one",
    "system-settings.backup-repo-cleanup.find-one": "system-settings.backup-repo-cleanup.find-one",
} as const;
