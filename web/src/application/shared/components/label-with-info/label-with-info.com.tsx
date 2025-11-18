import React from "react";

import classnames from "classnames/bind";

import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui";
import { Info } from "lucide-react";

import styles from "./label-with-info.module.scss";

const cx = classnames.bind(styles);

function View({ label, htmlFor, content, className, isRequired = false, icon }: LabelWithInfoProps) {
    return (
        <label
            className={cx("label-with-info", className)}
            htmlFor={htmlFor}
        >
            {icon && <span className={cx("icon")}>{icon}</span>}
            <span className={cx("label")}>{label} </span>
            {isRequired && <span className={cx("required")}>*</span>}

            {content && (
                <Tooltip>
                    <TooltipTrigger asChild>
                        <Info
                            size={14}
                            className="text-gray-500"
                        />
                    </TooltipTrigger>
                    <TooltipContent
                        side="right"
                        align="center"
                    >
                        {content}
                    </TooltipContent>
                </Tooltip>
            )}
        </label>
    );
}

export interface LabelWithInfoProps {
    label: React.ReactNode;
    htmlFor?: string;
    content?: React.ReactNode;
    className?: string;
    isRequired?: boolean;
    icon?: React.ReactNode;
}

export const LabelWithInfo = React.memo(View);
