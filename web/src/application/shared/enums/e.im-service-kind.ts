export const EImServiceKind = {
    Slack: "slack",
    Discord: "discord",
    Telegram: "telegram",
} as const;

export type EImServiceKind = (typeof EImServiceKind)[keyof typeof EImServiceKind];
