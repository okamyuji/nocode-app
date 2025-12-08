/**
 * ダッシュボードウィジェットグリッドコンポーネント
 * DnDでウィジェットを並び替え可能なグリッドレイアウト
 */

import { useDashboardWidgetsApi } from "@/api";
import type { DashboardWidget } from "@/types";
import { SimpleGrid, useToast } from "@chakra-ui/react";
import {
  closestCenter,
  DndContext,
  DragEndEvent,
  DragOverlay,
  DragStartEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import {
  arrayMove,
  rectSortingStrategy,
  SortableContext,
} from "@dnd-kit/sortable";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { DashboardWidgetCard } from "./DashboardWidgetCard";

interface DashboardWidgetGridProps {
  widgets: DashboardWidget[];
}

export function DashboardWidgetGrid({ widgets }: DashboardWidgetGridProps) {
  const toast = useToast();
  const queryClient = useQueryClient();
  const dashboardWidgetsApi = useDashboardWidgetsApi();

  const [activeId, setActiveId] = useState<number | null>(null);
  const [draggedItems, setDraggedItems] = useState<DashboardWidget[] | null>(
    null
  );

  // ドラッグ中のみdraggedItemsを使用、それ以外はpropsを直接使用
  // widgetsをメモ化して、内容が同じでも参照が変わった場合に再レンダリングを保証
  const items = useMemo(() => {
    if (draggedItems !== null) {
      return draggedItems;
    }
    return widgets;
  }, [draggedItems, widgets]);

  // DnDセンサー設定
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
    useSensor(KeyboardSensor)
  );

  // 並び替えミューテーション
  const reorderMutation = useMutation({
    mutationFn: (widgetIds: number[]) =>
      dashboardWidgetsApi.reorder({ widget_ids: widgetIds }),
    onSuccess: () => {
      // 成功時はドラッグステートをリセットして、新しいpropsを使用
      setDraggedItems(null);
      queryClient.invalidateQueries({ queryKey: ["dashboard", "widgets"] });
    },
    onError: () => {
      toast({
        title: "並び替えに失敗しました",
        status: "error",
        duration: 3000,
      });
      // エラー時はドラッグステートをリセット（propsに戻す）
      setDraggedItems(null);
    },
  });

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as number);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveId(null);

    if (over && active.id !== over.id) {
      const oldIndex = items.findIndex((item) => item.id === active.id);
      const newIndex = items.findIndex((item) => item.id === over.id);

      const newItems = arrayMove(items, oldIndex, newIndex);
      setDraggedItems(newItems);

      // APIに並び替えを送信
      reorderMutation.mutate(newItems.map((item) => item.id));
    }
  };

  const handleDragCancel = () => {
    setActiveId(null);
    setDraggedItems(null);
  };

  const activeWidget = activeId
    ? items.find((item) => item.id === activeId)
    : null;

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <SortableContext
        items={items.map((item) => item.id)}
        strategy={rectSortingStrategy}
      >
        <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={4}>
          {items.map((widget) => (
            <DashboardWidgetCard
              key={widget.id}
              widget={widget}
              isDragging={widget.id === activeId}
            />
          ))}
        </SimpleGrid>
      </SortableContext>

      <DragOverlay>
        {activeWidget ? (
          <DashboardWidgetCard widget={activeWidget} isDragging />
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
