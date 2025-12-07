/**
 * フィールドコード生成ユーティリティ
 * 外部データソースのカラム名から有効なフィールドコードを生成する
 */

/**
 * カラム名から有効なフィールドコードを生成する
 * - 英数字とアンダースコアのみ許可
 * - 英字で始まる必要がある
 * - 日本語など非ASCII文字はfield_{index}形式にフォールバック
 *
 * @param columnName 元のカラム名
 * @param index カラムのインデックス（フォールバック用）
 * @returns 有効なフィールドコード
 */
export function generateFieldCode(columnName: string, index: number): string {
  // 小文字に変換して、英数字とアンダースコア以外を除去
  let sanitized = columnName.toLowerCase().replace(/[^a-z0-9_]/g, "");

  // 末尾のアンダースコアを除去（"SPR2_プロセス" → "spr2_" → "spr2"）
  sanitized = sanitized.replace(/_+$/, "");

  // 空文字またはアンダースコアのみの場合はインデックスベースの名前を生成
  if (sanitized === "" || /^_+$/.test(sanitized)) {
    return `field_${index + 1}`;
  }

  // 数字またはアンダースコアで始まる場合はプレフィックスを付ける
  if (/^[0-9_]/.test(sanitized)) {
    return `f_${sanitized}`;
  }

  return sanitized;
}

/**
 * フィールドコードが有効かどうかを検証する
 *
 * 検証ルール:
 * - 英字（a-z, A-Z）で始まること
 * - 英数字とアンダースコアのみ使用可能
 * - 最大64文字（データベースカラム名の一般的な制限に準拠）
 *
 * @param code フィールドコード
 * @returns 有効な場合はtrue
 */
export function isValidFieldCode(code: string): boolean {
  if (!code || code.length > 64) return false;
  return /^[a-zA-Z][a-zA-Z0-9_]*$/.test(code);
}

/**
 * 重複を避けるためのユニークなフィールドコードを生成する
 *
 * @param columns カラム名の配列
 * @returns カラム名からユニークなフィールドコードへのマッピング
 * @throws 重複するカラム名がある場合はエラー
 */
export function generateUniqueFieldCodes(columns: { name: string }[]): {
  [key: string]: string;
} {
  // 重複するカラム名をチェック
  const columnNames = columns.map((c) => c.name);
  const duplicateNames = columnNames.filter(
    (name, index) => columnNames.indexOf(name) !== index
  );
  if (duplicateNames.length > 0) {
    const uniqueDuplicates = [...new Set(duplicateNames)];
    throw new Error(
      `重複するカラム名があります: ${uniqueDuplicates.join(", ")}`
    );
  }

  const usedCodes = new Set<string>();
  const result: { [key: string]: string } = {};

  columns.forEach((col, index) => {
    const code = generateFieldCode(col.name, index);
    let uniqueCode = code;
    let suffix = 1;

    // 生成されたコードが重複している場合はサフィックスを付ける
    // （異なるカラム名から同じコードが生成される場合、例: "名前" と "氏名" → 両方 "field_N"）
    while (usedCodes.has(uniqueCode)) {
      suffix++;
      uniqueCode = `${code}_${suffix}`;
    }

    usedCodes.add(uniqueCode);
    result[col.name] = uniqueCode;
  });

  return result;
}
