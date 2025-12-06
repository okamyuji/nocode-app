/**
 * アプリ状態管理ストア
 */

import { App, Field } from "@/types";
import { create } from "zustand";

/**
 * アプリ状態のインターフェース
 */
interface AppState {
  currentApp: App | null;
  currentFields: Field[];
  setCurrentApp: (app: App | null) => void;
  setCurrentFields: (fields: Field[]) => void;
  updateField: (field: Field) => void;
  addField: (field: Field) => void;
  removeField: (fieldId: number) => void;
  reorderFields: (fields: Field[]) => void;
}

export const useAppStore = create<AppState>((set, get) => ({
  currentApp: null,
  currentFields: [],

  // 現在のアプリを設定
  setCurrentApp: (app) => set({ currentApp: app }),

  // 現在のフィールドを設定（表示順でソート）
  setCurrentFields: (fields) =>
    set({
      currentFields: fields.sort((a, b) => a.display_order - b.display_order),
    }),

  // フィールドを更新
  updateField: (field) =>
    set({
      currentFields: get().currentFields.map((f) =>
        f.id === field.id ? field : f
      ),
    }),

  // フィールドを追加
  addField: (field) =>
    set({
      currentFields: [...get().currentFields, field].sort(
        (a, b) => a.display_order - b.display_order
      ),
    }),

  // フィールドを削除
  removeField: (fieldId) =>
    set({
      currentFields: get().currentFields.filter((f) => f.id !== fieldId),
    }),

  // フィールドの順序を変更
  reorderFields: (fields) => set({ currentFields: fields }),
}));
