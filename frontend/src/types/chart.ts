/**
 * チャート関連の型定義
 */

import { FilterItem } from "./record";

/**
 * チャート種類
 */
export type ChartType =
  | "bar"
  | "horizontal_bar"
  | "line"
  | "pie"
  | "doughnut"
  | "scatter"
  | "area";

/**
 * チャート軸
 */
export interface ChartAxis {
  field: string;
  label?: string;
  aggregation?: "count" | "sum" | "avg" | "min" | "max";
}

/**
 * チャートデータリクエスト
 */
export interface ChartDataRequest {
  chart_type: ChartType;
  x_axis: ChartAxis;
  y_axis: ChartAxis;
  filters?: FilterItem[];
  group_by?: string;
}

/**
 * チャートデータセット
 */
export interface ChartDataset {
  label: string;
  data: number[];
}

/**
 * チャートデータレスポンス
 */
export interface ChartDataResponse {
  labels: string[];
  datasets: ChartDataset[];
}

/**
 * チャート設定保存リクエスト
 */
export interface SaveChartConfigRequest {
  name: string;
  chart_type: ChartType;
  config: ChartDataRequest;
}

/**
 * チャート設定
 */
export interface ChartConfig {
  id: number;
  app_id: number;
  name: string;
  chart_type: ChartType;
  config: ChartDataRequest;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/**
 * チャート種類の表示ラベル
 */
export const CHART_TYPE_LABELS: Record<ChartType, string> = {
  bar: "棒グラフ",
  horizontal_bar: "横棒グラフ",
  line: "折れ線グラフ",
  pie: "円グラフ",
  doughnut: "ドーナツグラフ",
  scatter: "散布図",
  area: "面グラフ",
};
