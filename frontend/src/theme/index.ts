/**
 * Chakra UIテーマ設定
 */

import { createSystem, defaultConfig, defineConfig } from "@chakra-ui/react";
import { cardAnatomy, nativeSelectAnatomy } from "@chakra-ui/react/anatomy";

/**
 * カスタムカラー
 */
const colors = {
  brand: {
    50: { value: "#e6f3ff" },
    100: { value: "#b3d9ff" },
    200: { value: "#80bfff" },
    300: { value: "#4da6ff" },
    400: { value: "#1a8cff" },
    500: { value: "#0073e6" },
    600: { value: "#005ab3" },
    700: { value: "#004080" },
    800: { value: "#00264d" },
    900: { value: "#000d1a" },
  },
  accent: {
    50: { value: "#fff3e6" },
    100: { value: "#ffd9b3" },
    200: { value: "#ffbf80" },
    300: { value: "#ffa64d" },
    400: { value: "#ff8c1a" },
    500: { value: "#e67300" },
    600: { value: "#b35900" },
    700: { value: "#804000" },
    800: { value: "#4d2600" },
    900: { value: "#1a0d00" },
  },
};

// colorPalette="brand" / "accent" は、この意味トークンがないと色が付かない
const paletteSemanticTokens = (name: "brand" | "accent") => ({
  solid: { value: `{colors.${name}.500}` },
  contrast: { value: "white" },
  fg: { value: `{colors.${name}.700}` },
  muted: { value: `{colors.${name}.200}` },
  subtle: { value: `{colors.${name}.100}` },
  emphasized: { value: `{colors.${name}.300}` },
  focusRing: { value: `{colors.${name}.500}` },
});

/**
 * フォント設定
 */
const fonts = {
  heading: { value: '"Noto Sans JP", "Hiragino Sans", "Meiryo", sans-serif' },
  body: { value: '"Noto Sans JP", "Hiragino Sans", "Meiryo", sans-serif' },
};

const config = defineConfig({
  globalCss: {
    body: {
      bg: "gray.50",
      color: "gray.800",
    },
  },
  theme: {
    tokens: { colors, fonts },
    semanticTokens: {
      colors: {
        brand: paletteSemanticTokens("brand"),
        accent: paletteSemanticTokens("accent"),
      },
    },
    recipes: {
      // 既定の colorPalette は html の gray なので、ボタンと入力欄だけ brand にする
      button: {
        base: { colorPalette: "brand" },
        variants: {
          variant: {
            solid: { _hover: { bg: "colorPalette.600" } },
          },
        },
      },
      input: { base: { colorPalette: "brand" } },
      textarea: { base: { colorPalette: "brand" } },
    },
    slotRecipes: {
      nativeSelect: {
        slots: nativeSelectAnatomy.keys(),
        base: { field: { colorPalette: "brand" } },
      },
      card: {
        slots: cardAnatomy.keys(),
        base: { root: { borderRadius: "lg", boxShadow: "sm" } },
        defaultVariants: { variant: "elevated" },
      },
    },
  },
});

export const system = createSystem(defaultConfig, config);
