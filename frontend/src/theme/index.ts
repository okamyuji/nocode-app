/**
 * Chakra UIテーマ設定
 */

import { extendTheme, type ThemeConfig } from "@chakra-ui/react";

/**
 * カラーモード設定
 */
const config: ThemeConfig = {
  initialColorMode: "light",
  useSystemColorMode: false,
};

/**
 * カスタムカラー
 */
const colors = {
  brand: {
    50: "#e6f3ff",
    100: "#b3d9ff",
    200: "#80bfff",
    300: "#4da6ff",
    400: "#1a8cff",
    500: "#0073e6",
    600: "#005ab3",
    700: "#004080",
    800: "#00264d",
    900: "#000d1a",
  },
  accent: {
    50: "#fff3e6",
    100: "#ffd9b3",
    200: "#ffbf80",
    300: "#ffa64d",
    400: "#ff8c1a",
    500: "#e67300",
    600: "#b35900",
    700: "#804000",
    800: "#4d2600",
    900: "#1a0d00",
  },
};

/**
 * フォント設定
 */
const fonts = {
  heading: '"Noto Sans JP", "Hiragino Sans", "Meiryo", sans-serif',
  body: '"Noto Sans JP", "Hiragino Sans", "Meiryo", sans-serif',
};

/**
 * コンポーネント設定
 */
const components = {
  Button: {
    defaultProps: {
      colorScheme: "brand",
    },
    variants: {
      solid: {
        bg: "brand.500",
        color: "white",
        _hover: {
          bg: "brand.600",
        },
      },
    },
  },
  Input: {
    defaultProps: {
      focusBorderColor: "brand.500",
    },
  },
  Select: {
    defaultProps: {
      focusBorderColor: "brand.500",
    },
  },
  Textarea: {
    defaultProps: {
      focusBorderColor: "brand.500",
    },
  },
  Card: {
    baseStyle: {
      container: {
        borderRadius: "lg",
        boxShadow: "sm",
      },
    },
  },
};

/**
 * グローバルスタイル
 */
const styles = {
  global: {
    body: {
      bg: "gray.50",
      color: "gray.800",
    },
  },
};

export const theme = extendTheme({
  config,
  colors,
  fonts,
  components,
  styles,
});
