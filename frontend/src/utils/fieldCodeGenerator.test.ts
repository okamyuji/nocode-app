/**
 * フィールドコード生成ユーティリティのテスト
 */

import { describe, expect, it } from "vitest";
import {
  generateFieldCode,
  generateUniqueFieldCodes,
  isValidFieldCode,
} from "./fieldCodeGenerator";

describe("generateFieldCode", () => {
  describe("英語カラム名", () => {
    it("シンプルな英語名はそのまま返す", () => {
      expect(generateFieldCode("users", 0)).toBe("users");
      expect(generateFieldCode("category", 1)).toBe("category");
      expect(generateFieldCode("amount", 2)).toBe("amount");
    });

    it("大文字は小文字に変換する", () => {
      expect(generateFieldCode("Users", 0)).toBe("users");
      expect(generateFieldCode("CATEGORY", 1)).toBe("category");
      expect(generateFieldCode("AmounT", 2)).toBe("amount");
    });

    it("アンダースコアは保持する", () => {
      expect(generateFieldCode("user_name", 0)).toBe("user_name");
      expect(generateFieldCode("first_name", 1)).toBe("first_name");
      expect(generateFieldCode("created_at", 2)).toBe("created_at");
    });

    it("数字で始まる場合はプレフィックスを付ける", () => {
      expect(generateFieldCode("123abc", 0)).toBe("f_123abc");
      expect(generateFieldCode("1column", 1)).toBe("f_1column");
    });
  });

  describe("日本語カラム名", () => {
    it("全て日本語の場合はfield_{index}形式を返す", () => {
      expect(generateFieldCode("プロセス名", 0)).toBe("field_1");
      expect(generateFieldCode("顧客名", 1)).toBe("field_2");
      expect(generateFieldCode("売上金額", 2)).toBe("field_3");
    });

    it("ひらがな・カタカナ・漢字すべてが除去される", () => {
      expect(generateFieldCode("あいうえお", 0)).toBe("field_1");
      expect(generateFieldCode("アイウエオ", 1)).toBe("field_2");
      expect(generateFieldCode("日本語", 2)).toBe("field_3");
    });
  });

  describe("混合カラム名（英語+日本語）", () => {
    it("英数字部分のみを抽出し、末尾アンダースコアは除去する", () => {
      expect(generateFieldCode("SPR2_プロセスマスタ", 0)).toBe("spr2");
      expect(generateFieldCode("user_名前", 1)).toBe("user");
      expect(generateFieldCode("spr2_プロセス_マスタ", 2)).toBe("spr2");
    });

    it("全角数字は除去される", () => {
      expect(generateFieldCode("テスト１２３", 0)).toBe("field_1");
    });

    it("混合で数字で始まる場合もプレフィックスを付ける", () => {
      expect(generateFieldCode("1テスト", 0)).toBe("f_1");
    });
  });

  describe("特殊文字", () => {
    it("スペースは除去される", () => {
      expect(generateFieldCode("user name", 0)).toBe("username");
      expect(generateFieldCode("first name", 1)).toBe("firstname");
    });

    it("ハイフンは除去される", () => {
      expect(generateFieldCode("user-name", 0)).toBe("username");
    });

    it("ドットは除去される", () => {
      expect(generateFieldCode("user.name", 0)).toBe("username");
    });

    it("特殊記号は除去される", () => {
      expect(generateFieldCode("user@name", 0)).toBe("username");
      expect(generateFieldCode("user#name", 1)).toBe("username");
      expect(generateFieldCode("user$name", 2)).toBe("username");
    });
  });

  describe("エッジケース", () => {
    it("空文字はfield_{index}形式を返す", () => {
      expect(generateFieldCode("", 0)).toBe("field_1");
      expect(generateFieldCode("", 5)).toBe("field_6");
    });

    it("記号のみの場合はfield_{index}形式を返す", () => {
      expect(generateFieldCode("@#$%", 0)).toBe("field_1");
      expect(generateFieldCode("!!!!", 1)).toBe("field_2");
    });

    it("アンダースコアのみの場合はfield_{index}形式を返す", () => {
      expect(generateFieldCode("___", 0)).toBe("field_1");
      expect(generateFieldCode("_", 1)).toBe("field_2");
    });

    it("末尾アンダースコアは除去される", () => {
      expect(generateFieldCode("test_", 0)).toBe("test");
      expect(generateFieldCode("test___", 1)).toBe("test");
    });
  });
});

describe("isValidFieldCode", () => {
  describe("有効なフィールドコード", () => {
    it("英字で始まり英数字のみの場合はtrue", () => {
      expect(isValidFieldCode("users")).toBe(true);
      expect(isValidFieldCode("user123")).toBe(true);
      expect(isValidFieldCode("a")).toBe(true);
    });

    it("アンダースコアを含む場合もtrue", () => {
      expect(isValidFieldCode("user_name")).toBe(true);
      expect(isValidFieldCode("first_name_123")).toBe(true);
      expect(isValidFieldCode("a_b_c")).toBe(true);
    });

    it("大文字を含む場合もtrue", () => {
      expect(isValidFieldCode("Users")).toBe(true);
      expect(isValidFieldCode("UserName")).toBe(true);
      expect(isValidFieldCode("ABC123")).toBe(true);
    });

    it("field_N形式もtrue", () => {
      expect(isValidFieldCode("field_1")).toBe(true);
      expect(isValidFieldCode("field_123")).toBe(true);
    });
  });

  describe("無効なフィールドコード", () => {
    it("空文字はfalse", () => {
      expect(isValidFieldCode("")).toBe(false);
    });

    it("数字で始まる場合はfalse", () => {
      expect(isValidFieldCode("123abc")).toBe(false);
      expect(isValidFieldCode("1user")).toBe(false);
    });

    it("日本語を含む場合はfalse", () => {
      expect(isValidFieldCode("userプロセス")).toBe(false);
      expect(isValidFieldCode("名前")).toBe(false);
    });

    it("特殊文字を含む場合はfalse", () => {
      expect(isValidFieldCode("user-name")).toBe(false);
      expect(isValidFieldCode("user.name")).toBe(false);
      expect(isValidFieldCode("user@name")).toBe(false);
    });

    it("スペースを含む場合はfalse", () => {
      expect(isValidFieldCode("user name")).toBe(false);
    });

    it("64文字を超える場合はfalse", () => {
      const longCode = "a".repeat(65);
      expect(isValidFieldCode(longCode)).toBe(false);
    });

    it("アンダースコアで始まる場合はfalse", () => {
      expect(isValidFieldCode("_user")).toBe(false);
      expect(isValidFieldCode("__test")).toBe(false);
    });
  });
});

describe("generateUniqueFieldCodes", () => {
  describe("重複なしの場合", () => {
    it("各カラムにユニークなコードを割り当てる", () => {
      const columns = [{ name: "id" }, { name: "name" }, { name: "email" }];
      const result = generateUniqueFieldCodes(columns);

      expect(result["id"]).toBe("id");
      expect(result["name"]).toBe("name");
      expect(result["email"]).toBe("email");
    });

    it("日本語カラムにはfield_N形式を割り当てる", () => {
      const columns = [
        { name: "プロセスコード" },
        { name: "プロセス名" },
        { name: "金額" },
      ];
      const result = generateUniqueFieldCodes(columns);

      expect(result["プロセスコード"]).toBe("field_1");
      expect(result["プロセス名"]).toBe("field_2");
      expect(result["金額"]).toBe("field_3");
    });
  });

  describe("重複カラム名がある場合", () => {
    it("同じカラム名がある場合はエラーを投げる", () => {
      const columns = [{ name: "name" }, { name: "name" }, { name: "name" }];
      expect(() => generateUniqueFieldCodes(columns)).toThrow(
        "重複するカラム名があります: name"
      );
    });

    it("日本語が同名の場合もエラーを投げる", () => {
      const columns = [{ name: "名前" }, { name: "名前" }];
      expect(() => generateUniqueFieldCodes(columns)).toThrow(
        "重複するカラム名があります: 名前"
      );
    });

    it("複数の重複カラムがある場合はすべて表示する", () => {
      const columns = [
        { name: "name" },
        { name: "email" },
        { name: "name" },
        { name: "email" },
      ];
      expect(() => generateUniqueFieldCodes(columns)).toThrow(
        "重複するカラム名があります: name, email"
      );
    });
  });

  describe("生成されたコードが重複する場合", () => {
    it("異なるカラム名から同じコードが生成される場合はサフィックスを付ける", () => {
      // "名前" と "氏名" は両方 field_N 形式になるが、indexが異なるので別の値になる
      // "名前" (index=0) → field_1, "氏名" (index=1) → field_2
      const columns = [{ name: "名前" }, { name: "氏名" }];
      const result = generateUniqueFieldCodes(columns);

      expect(result["名前"]).toBe("field_1");
      expect(result["氏名"]).toBe("field_2");
    });

    it("同じbaseコードが生成される場合はサフィックスを付ける", () => {
      // 同じindex=0のbaseコードが重複した場合のテスト
      // "test" と "TEST" は両方 "test" になる→2番目にサフィックス
      const columns = [{ name: "test" }, { name: "TEST" }];
      const result = generateUniqueFieldCodes(columns);

      expect(result["test"]).toBe("test");
      expect(result["TEST"]).toBe("test_2");
    });

    it("英語変換で同じコードになる場合もサフィックスを付ける", () => {
      // "User Name" と "user_name" は両方 "username" になる
      const columns = [{ name: "User Name" }, { name: "user-name" }];
      const result = generateUniqueFieldCodes(columns);

      expect(result["User Name"]).toBe("username");
      expect(result["user-name"]).toBe("username_2");
    });
  });

  describe("混合ケース", () => {
    it("英語と日本語が混在する場合", () => {
      const columns = [
        { name: "id" },
        { name: "プロセスコード" },
        { name: "name" },
        { name: "プロセス名" },
      ];
      const result = generateUniqueFieldCodes(columns);

      expect(result["id"]).toBe("id");
      expect(result["プロセスコード"]).toBe("field_2");
      expect(result["name"]).toBe("name");
      expect(result["プロセス名"]).toBe("field_4");
    });

    it("Oracle形式の混合名（英語+日本語）末尾アンダースコアは除去", () => {
      const columns = [
        { name: "SPR2_プロセスマスタ" },
        { name: "CODE_コード" },
      ];
      const result = generateUniqueFieldCodes(columns);

      expect(result["SPR2_プロセスマスタ"]).toBe("spr2");
      expect(result["CODE_コード"]).toBe("code");
    });
  });

  describe("エッジケース", () => {
    it("空配列の場合は空オブジェクトを返す", () => {
      const result = generateUniqueFieldCodes([]);
      expect(result).toEqual({});
    });

    it("1つだけのカラム", () => {
      const result = generateUniqueFieldCodes([{ name: "single" }]);
      expect(result["single"]).toBe("single");
    });
  });
});
