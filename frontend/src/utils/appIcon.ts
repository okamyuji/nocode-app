/**
 * アプリアイコンユーティリティ
 * アプリのアイコン名からReact Iconコンポーネントを返す
 */

import { IconType } from "react-icons";
import { FiCalendar, FiDatabase, FiGrid, FiList } from "react-icons/fi";

/**
 * アイコン名からReact Iconコンポーネントを取得する
 *
 * @param iconName アイコン名（"default", "grid", "list", "calendar", "database"）
 * @returns 対応するReact Iconコンポーネント
 */
export function getAppIcon(iconName: string | undefined): IconType {
  switch (iconName) {
    case "grid":
      return FiGrid;
    case "list":
      return FiList;
    case "calendar":
      return FiCalendar;
    case "database":
      return FiDatabase;
    case "default":
    default:
      return FiGrid;
  }
}
