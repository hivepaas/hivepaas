import JavascriptTimeAgo from "javascript-time-ago";
import en from "javascript-time-ago/locale/en";

JavascriptTimeAgo.addDefaultLocale(en);

export const timeAgoFormatter = new JavascriptTimeAgo("en-US");
