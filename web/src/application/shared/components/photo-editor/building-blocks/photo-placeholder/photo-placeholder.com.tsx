import React from "react";

import classnames from "classnames/bind";

import { type LucideIcon, SquareUserRound } from "lucide-react";

import styles from "./photo-placeholder.module.scss";

const cx = classnames.bind(styles);

function View({ hideIcon, icon }: Props) {
    const IconComponent = icon ?? SquareUserRound;

    return <div className={cx("photo-placeholder")}>{!hideIcon && <IconComponent className={cx("icon")} />}</div>;
}

export const PhotoPlaceholder = React.memo(View);

interface Props {
    hideIcon?: boolean;
    icon?: LucideIcon;
}
