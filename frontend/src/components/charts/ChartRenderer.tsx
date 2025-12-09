import { ChartDataResponse, ChartType } from "@/types";
import { Box } from "@chakra-ui/react";
import { useMemo } from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Scatter,
  ScatterChart,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

interface ChartRendererProps {
  data: ChartDataResponse;
  chartType: ChartType;
  height?: number | string;
}

const COLORS = [
  "#0073e6",
  "#e67300",
  "#00b894",
  "#e74c3c",
  "#9b59b6",
  "#3498db",
  "#2ecc71",
  "#f39c12",
  "#1abc9c",
  "#e91e63",
];

export function ChartRenderer({
  data,
  chartType,
  height = 400,
}: ChartRendererProps) {
  const chartData = useMemo(() => {
    if (!data || !data.labels) return [];

    return data.labels.map((label, index) => {
      const item: Record<string, string | number> = { name: label };
      data.datasets.forEach((dataset) => {
        item[dataset.label] = dataset.data[index] || 0;
      });
      return item;
    });
  }, [data]);

  const pieData = useMemo(() => {
    if (
      !data ||
      !data.labels ||
      (chartType !== "pie" && chartType !== "doughnut")
    ) {
      return [];
    }

    return data.labels.map((label, index) => ({
      name: label,
      value: data.datasets[0]?.data[index] || 0,
    }));
  }, [data, chartType]);

  const renderChart = () => {
    switch (chartType) {
      case "bar":
        return (
          <BarChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="name" />
            <YAxis />
            <Tooltip />
            <Legend />
            {data.datasets.map((dataset, index) => (
              <Bar
                key={dataset.label}
                dataKey={dataset.label}
                fill={COLORS[index % COLORS.length]}
              />
            ))}
          </BarChart>
        );

      case "horizontal_bar":
        return (
          <BarChart data={chartData} layout="vertical">
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis type="number" />
            <YAxis dataKey="name" type="category" width={100} />
            <Tooltip />
            <Legend />
            {data.datasets.map((dataset, index) => (
              <Bar
                key={dataset.label}
                dataKey={dataset.label}
                fill={COLORS[index % COLORS.length]}
              />
            ))}
          </BarChart>
        );

      case "line":
        return (
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="name" />
            <YAxis />
            <Tooltip />
            <Legend />
            {data.datasets.map((dataset, index) => (
              <Line
                key={dataset.label}
                type="monotone"
                dataKey={dataset.label}
                stroke={COLORS[index % COLORS.length]}
                strokeWidth={2}
              />
            ))}
          </LineChart>
        );

      case "area":
        return (
          <AreaChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="name" />
            <YAxis />
            <Tooltip />
            <Legend />
            {data.datasets.map((dataset, index) => (
              <Area
                key={dataset.label}
                type="monotone"
                dataKey={dataset.label}
                stroke={COLORS[index % COLORS.length]}
                fill={COLORS[index % COLORS.length]}
                fillOpacity={0.3}
              />
            ))}
          </AreaChart>
        );

      case "pie":
      case "doughnut":
        return (
          <PieChart>
            <Pie
              data={pieData}
              cx="50%"
              cy="50%"
              innerRadius={chartType === "doughnut" ? 60 : 0}
              outerRadius={120}
              dataKey="value"
              label={({ name, percent }) =>
                `${name} (${(percent * 100).toFixed(0)}%)`
              }
            >
              {pieData.map((_, index) => (
                <Cell
                  key={`cell-${index}`}
                  fill={COLORS[index % COLORS.length]}
                />
              ))}
            </Pie>
            <Tooltip />
            <Legend />
          </PieChart>
        );

      case "scatter":
        return (
          <ScatterChart>
            <CartesianGrid />
            <XAxis dataKey="x" type="number" name="X" />
            <YAxis dataKey="y" type="number" name="Y" />
            <Tooltip cursor={{ strokeDasharray: "3 3" }} />
            <Legend />
            <Scatter
              name={data.datasets[0]?.label || "データ"}
              data={chartData.map((item, index) => ({
                x: index,
                y: item[data.datasets[0]?.label || "value"] as number,
              }))}
              fill={COLORS[0]}
            />
          </ScatterChart>
        );

      default:
        return (
          <BarChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="name" />
            <YAxis />
            <Tooltip />
            <Legend />
            {data.datasets.map((dataset, index) => (
              <Bar
                key={dataset.label}
                dataKey={dataset.label}
                fill={COLORS[index % COLORS.length]}
              />
            ))}
          </BarChart>
        );
    }
  };

  return (
    <Box w="100%" h={height}>
      <ResponsiveContainer width="100%" height="100%">
        {renderChart()}
      </ResponsiveContainer>
    </Box>
  );
}
