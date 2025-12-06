/**
 * ローディングコンポーネントのテスト
 */

import { render, screen } from "@/test/utils";
import { describe, expect, it } from "vitest";
import { Loading } from "./Loading";

describe("Loading", () => {
  it("デフォルトメッセージで表示される", () => {
    render(<Loading />);

    expect(screen.getByText("読み込み中...")).toBeInTheDocument();
  });

  it("カスタムメッセージで表示される", () => {
    render(<Loading message="データを取得中..." />);

    expect(screen.getByText("データを取得中...")).toBeInTheDocument();
  });

  it("スピナーが表示される", () => {
    render(<Loading />);

    expect(screen.getByText("読み込み中...")).toBeInTheDocument();
  });

  it("デフォルトでは非フルスクリーンモードで表示される", () => {
    const { container } = render(<Loading />);

    // 非フルスクリーンモードではposition: fixedを持たない
    const outerElement = container.firstChild as HTMLElement;
    expect(outerElement).not.toHaveStyle({ position: "fixed" });
  });

  it("指定時はフルスクリーンモードで表示される", () => {
    const { container } = render(<Loading fullScreen />);

    // フルスクリーンモードではposition: fixedを持つ
    const outerElement = container.firstChild as HTMLElement;
    expect(outerElement).toHaveStyle({ position: "fixed" });
  });
});
