import React, { type PropsWithChildren } from "react";

import classnames from "classnames/bind";

import styles from "./info-block.module.scss";

const cx = classnames.bind(styles);

function View({ title, description, children, titleWidth = 270 }: Props) {
    return (
        <div className={cx("info-block")}>
            <div className={cx("info")} style={{ minWidth: titleWidth }}>
                <div className={cx("title")}>{title}</div>

                {description && <div className={cx("description")}>{description}</div>}
            </div>

            <div className={cx("children")}>{children}</div>
        </div>
    );
}

type Props = PropsWithChildren<{
    title: React.ReactNode;
    description?: React.ReactNode;
    titleWidth?: number;
}>;

export const InfoBlock = React.memo(View);
