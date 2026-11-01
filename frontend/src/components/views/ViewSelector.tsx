import { ViewType } from "@/types";
import { CalendarIcon } from "@chakra-ui/icons/Calendar";
import { Tooltip } from "@/components/ui/tooltip";
import { Button, ButtonGroup, Icon } from "@chakra-ui/react";
import { FiBarChart2, FiGrid, FiList } from "react-icons/fi";

interface ViewSelectorProps {
  currentView: ViewType;
  onViewChange: (view: ViewType) => void;
  hasDateField?: boolean;
}

const viewOptions: {
  type: ViewType;
  label: string;
  icon: React.ReactNode;
  tooltip: string;
}[] = [
  {
    type: "table",
    label: "テーブル",
    icon: <Icon as={FiGrid} />,
    tooltip: "テーブル形式で表示",
  },
  {
    type: "list",
    label: "リスト",
    icon: <Icon as={FiList} />,
    tooltip: "カード形式で表示",
  },
  {
    type: "calendar",
    label: "カレンダー",
    icon: <CalendarIcon />,
    tooltip: "カレンダー形式で表示",
  },
  {
    type: "chart",
    label: "グラフ",
    icon: <Icon as={FiBarChart2} />,
    tooltip: "グラフで可視化",
  },
];

export function ViewSelector({
  currentView,
  onViewChange,
  hasDateField = true,
}: ViewSelectorProps) {
  return (
    <ButtonGroup size="sm" attached variant="outline">
      {viewOptions.map((option) => {
        // Hide calendar view if no date field
        if (option.type === "calendar" && !hasDateField) {
          return null;
        }

        return (
          <Tooltip key={option.type} content={option.tooltip}>
            <Button
              aria-pressed={currentView === option.type}
              onClick={() => onViewChange(option.type)}
              colorPalette={currentView === option.type ? "brand" : "gray"}
              variant={currentView === option.type ? "solid" : "outline"}
            >
              {option.icon}
              {option.label}
            </Button>
          </Tooltip>
        );
      })}
    </ButtonGroup>
  );
}
