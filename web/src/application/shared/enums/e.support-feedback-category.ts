export const SupportFeedbackCategory = {
    General: "general",
    SecurityReport: "security_report",
    BugIssueReport: "bug_issue_report",
    Licensing: "licensing",
} as const;

export type SupportFeedbackCategory = (typeof SupportFeedbackCategory)[keyof typeof SupportFeedbackCategory];
