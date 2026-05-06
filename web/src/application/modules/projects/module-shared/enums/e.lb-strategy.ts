export const ELBStrategy = {
    WeightedRoundRobin: "wrr",
    PowerOfTwoChoices: "p2c",
    HighestRandomWeight: "hrw",
    LeastTime: "leasttime",
} as const;

export type ELBStrategy = (typeof ELBStrategy)[keyof typeof ELBStrategy];
