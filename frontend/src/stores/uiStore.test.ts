import { beforeEach, describe, expect, it } from "vitest";
import {
  DEFAULT_SIDEBAR_WIDTH,
  MAX_SIDEBAR_WIDTH,
  MIN_SIDEBAR_WIDTH,
  useUIStore,
} from "./uiStore";

describe("uiStore", () => {
  beforeEach(() => {
    // Reset store state before each test
    useUIStore.setState({
      sidebarWidth: DEFAULT_SIDEBAR_WIDTH,
      sidebarCollapsed: false,
    });
    localStorage.clear();
  });

  describe("initial state", () => {
    it("should have default values", () => {
      const state = useUIStore.getState();
      expect(state.sidebarWidth).toBe(DEFAULT_SIDEBAR_WIDTH);
      expect(state.sidebarCollapsed).toBe(false);
    });
  });

  describe("setSidebarWidth", () => {
    it("should set sidebar width", () => {
      useUIStore.getState().setSidebarWidth(300);

      expect(useUIStore.getState().sidebarWidth).toBe(300);
    });

    it("should clamp width to minimum", () => {
      useUIStore.getState().setSidebarWidth(10);

      expect(useUIStore.getState().sidebarWidth).toBe(MIN_SIDEBAR_WIDTH);
    });

    it("should clamp width to maximum", () => {
      useUIStore.getState().setSidebarWidth(1000);

      expect(useUIStore.getState().sidebarWidth).toBe(MAX_SIDEBAR_WIDTH);
    });

    it("should accept boundary values", () => {
      useUIStore.getState().setSidebarWidth(MIN_SIDEBAR_WIDTH);
      expect(useUIStore.getState().sidebarWidth).toBe(MIN_SIDEBAR_WIDTH);

      useUIStore.getState().setSidebarWidth(MAX_SIDEBAR_WIDTH);
      expect(useUIStore.getState().sidebarWidth).toBe(MAX_SIDEBAR_WIDTH);
    });
  });

  describe("toggleSidebarCollapsed", () => {
    it("should toggle collapsed state from false to true", () => {
      expect(useUIStore.getState().sidebarCollapsed).toBe(false);

      useUIStore.getState().toggleSidebarCollapsed();

      expect(useUIStore.getState().sidebarCollapsed).toBe(true);
    });

    it("should toggle collapsed state from true to false", () => {
      useUIStore.setState({ sidebarCollapsed: true });

      useUIStore.getState().toggleSidebarCollapsed();

      expect(useUIStore.getState().sidebarCollapsed).toBe(false);
    });

    it("should toggle multiple times correctly", () => {
      useUIStore.getState().toggleSidebarCollapsed();
      expect(useUIStore.getState().sidebarCollapsed).toBe(true);

      useUIStore.getState().toggleSidebarCollapsed();
      expect(useUIStore.getState().sidebarCollapsed).toBe(false);

      useUIStore.getState().toggleSidebarCollapsed();
      expect(useUIStore.getState().sidebarCollapsed).toBe(true);
    });
  });

  describe("exported constants", () => {
    it("should export correct constant values", () => {
      expect(MIN_SIDEBAR_WIDTH).toBe(60);
      expect(MAX_SIDEBAR_WIDTH).toBe(400);
      expect(DEFAULT_SIDEBAR_WIDTH).toBe(240);
    });
  });
});
