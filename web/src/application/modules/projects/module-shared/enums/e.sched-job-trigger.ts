export const ESchedJobTriggerEvent = {
    PreDeploy: "pre-deploy",
    PostDeploy: "post-deploy",
    DeployFailed: "deploy-failed",
    HealthDown: "health-down",
    HealthUp: "health-up",
    AppEnabled: "app-enabled",
    AppDisabled: "app-disabled",
} as const;

export type ESchedJobTriggerEvent = (typeof ESchedJobTriggerEvent)[keyof typeof ESchedJobTriggerEvent];

/** The events in the order a form offers them, with their labels. */
export const SCHED_JOB_TRIGGER_EVENT_OPTIONS: { value: ESchedJobTriggerEvent; label: string }[] = [
    { value: ESchedJobTriggerEvent.PreDeploy, label: "Before a deploy" },
    { value: ESchedJobTriggerEvent.PostDeploy, label: "After a deploy" },
    { value: ESchedJobTriggerEvent.DeployFailed, label: "A deploy failed" },
    { value: ESchedJobTriggerEvent.HealthDown, label: "Health check down" },
    { value: ESchedJobTriggerEvent.HealthUp, label: "Health check up again" },
    { value: ESchedJobTriggerEvent.AppEnabled, label: "App enabled" },
    { value: ESchedJobTriggerEvent.AppDisabled, label: "App disabled" },
];
