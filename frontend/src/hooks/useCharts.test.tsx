import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  useChartConfigs,
  useChartData,
  useDeleteChartConfig,
  useSaveChartConfig,
} from "./useCharts";

// chartsApiのモック
vi.mock("@/api", () => ({
  chartsApi: {
    getData: vi.fn(),
    getConfigs: vi.fn(),
    saveConfig: vi.fn(),
    deleteConfig: vi.fn(),
  },
}));

import { chartsApi } from "@/api";
import { ChartConfig, ChartDataResponse, ChartType } from "@/types";

const mockChartDataResponse: ChartDataResponse = {
  labels: ["Category A", "Category B"],
  datasets: [{ label: "Amount", data: [10, 20] }],
};

const mockChartConfig: ChartConfig = {
  id: 1,
  app_id: 1,
  name: "Test Chart",
  chart_type: "bar" as ChartType,
  config: {
    chart_type: "bar" as ChartType,
    x_axis: { field: "category" },
    y_axis: { field: "amount", aggregation: "sum" },
  },
  created_by: 1,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

const mockChartConfigsResponse = {
  configs: [mockChartConfig],
};

describe("useCharts hooks", () => {
  let queryClient: QueryClient;

  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          gcTime: 0,
        },
        mutations: {
          retry: false,
        },
      },
    });
    vi.clearAllMocks();
  });

  describe("useChartData", () => {
    it("should fetch chart data when request is provided", async () => {
      vi.mocked(chartsApi.getData).mockResolvedValueOnce(mockChartDataResponse);

      const request = {
        chart_type: "bar" as const,
        x_axis: { field: "category" },
        y_axis: { field: "amount", aggregation: "sum" as const },
      };

      const { result } = renderHook(() => useChartData(1, request), {
        wrapper,
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(chartsApi.getData).toHaveBeenCalledWith(1, request);
      expect(result.current.data).toEqual(mockChartDataResponse);
    });

    it("should not fetch when appId is 0", () => {
      const request = {
        chart_type: "bar" as const,
        x_axis: { field: "category" },
        y_axis: { field: "amount", aggregation: "sum" as const },
      };

      const { result } = renderHook(() => useChartData(0, request), {
        wrapper,
      });

      expect(result.current.isFetching).toBe(false);
      expect(chartsApi.getData).not.toHaveBeenCalled();
    });

    it("should not fetch when request is null", () => {
      const { result } = renderHook(() => useChartData(1, null), { wrapper });

      expect(result.current.isFetching).toBe(false);
      expect(chartsApi.getData).not.toHaveBeenCalled();
    });
  });

  describe("useChartConfigs", () => {
    it("should fetch chart configs", async () => {
      vi.mocked(chartsApi.getConfigs).mockResolvedValueOnce(
        mockChartConfigsResponse
      );

      const { result } = renderHook(() => useChartConfigs(1), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(chartsApi.getConfigs).toHaveBeenCalledWith(1);
      expect(result.current.data).toEqual(mockChartConfigsResponse);
    });

    it("should not fetch when appId is 0", () => {
      const { result } = renderHook(() => useChartConfigs(0), { wrapper });

      expect(result.current.isFetching).toBe(false);
      expect(chartsApi.getConfigs).not.toHaveBeenCalled();
    });
  });

  describe("useSaveChartConfig", () => {
    it("should save chart config and invalidate queries", async () => {
      vi.mocked(chartsApi.saveConfig).mockResolvedValueOnce(mockChartConfig);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useSaveChartConfig(), { wrapper });

      result.current.mutate({
        appId: 1,
        data: {
          name: "Test Chart",
          chart_type: "bar",
          config: {
            chart_type: "bar",
            x_axis: { field: "category" },
            y_axis: { field: "amount", aggregation: "sum" },
          },
        },
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(chartsApi.saveConfig).toHaveBeenCalled();
      expect(invalidateSpy).toHaveBeenCalledWith({
        queryKey: ["chartConfigs", 1],
      });
    });
  });

  describe("useDeleteChartConfig", () => {
    it("should delete chart config and invalidate queries", async () => {
      vi.mocked(chartsApi.deleteConfig).mockResolvedValueOnce(undefined);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useDeleteChartConfig(), { wrapper });

      result.current.mutate({ appId: 1, configId: 1 });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(chartsApi.deleteConfig).toHaveBeenCalledWith(1, 1);
      expect(invalidateSpy).toHaveBeenCalledWith({
        queryKey: ["chartConfigs", 1],
      });
    });
  });
});
