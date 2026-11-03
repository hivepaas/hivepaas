import type { BadgeTone } from "@components/ui/badge";

/**
 * The color of a messaging channel, wherever one is named: an IM platform and
 * the notification target that sends through it show the same badge.
 */
export const CHANNEL_TONES: Record<string, BadgeTone> = {
    email: "blue",
    slack: "purple",
    discord: "indigo",
    telegram: "sky",
    lark: "teal",
};

/** A channel added later, until it is given a color of its own. */
export const OTHER_CHANNEL_TONE: BadgeTone = "amber";

export function channelTone(channel: string): BadgeTone {
    return CHANNEL_TONES[channel.toLowerCase()] ?? OTHER_CHANNEL_TONE;
}
