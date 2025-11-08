import { useDeferredValue, useMemo } from "react";

import classnames from "classnames/bind";

import { useMount, useUpdateEffect } from "react-use";

import { CheckCircleIcon } from "@assets/icons";

import styles from "./password-strength-meter.module.scss";

const cx = classnames.bind(styles);

type Strength = number | "max";

interface Rule {
    title: string;
    check: (input: string) => boolean;
}

const specialCharacters = /[!"#$%&'()*+,-./:;<=>?@[\]^_`{|}~]/;

const rules: Rule[] = [
    {
        title: "one lowercase character",
        check: (input: string): boolean => {
            return input.toUpperCase() !== input;
        },
    },
    {
        title: "one special character",
        check: (input: string): boolean => {
            return specialCharacters.test(input);
        },
    },
    {
        title: "one uppercase character",
        check: (input: string): boolean => {
            return input.toLowerCase() !== input;
        },
    },
    {
        title: "8 character minimum",
        check: (input: string): boolean => {
            return input.trim().length >= 8;
        },
    },
    {
        title: "one number",
        check: (input: string): boolean => {
            return /[0-9]/.test(input);
        },
    },
];

export function PasswordStrengthMeter({ password, onStrengthChange }: Props) {
    const deferred = useDeferredValue(password);

    const strength: Strength = useMemo(() => {
        if (deferred.length === 0) {
            return 0;
        }

        const successCount = rules.reduce((acc, rule) => {
            return acc + (rule.check(deferred) ? 1 : 0);
        }, 0);

        return successCount === rules.length ? "max" : successCount;
    }, [deferred]);

    useMount(() => {
        onStrengthChange(strength);
    });

    useUpdateEffect(() => {
        onStrengthChange(strength);
    }, [strength]);

    return (
        <div className={cx("password-strength-meter")}>
            {rules.map(rule => {
                const success = rule.check(deferred);

                return (
                    <div
                        key={rule.title}
                        className={cx("rule", {
                            success,
                        })}
                    >
                        <div className={cx("icon")}>
                            <CheckCircleIcon />
                        </div>
                        <div className={cx("title")}>{rule.title}</div>
                    </div>
                );
            })}
        </div>
    );
}

interface Props {
    password: string;
    onStrengthChange: (strength: Strength) => void;
}
