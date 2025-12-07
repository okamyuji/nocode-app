import { App, Field, FieldType } from "@/types";
import { beforeEach, describe, expect, it } from "vitest";
import { useAppStore } from "./appStore";

describe("appStore", () => {
  const mockApp: App = {
    id: 1,
    name: "Test App",
    description: "A test application",
    table_name: "app_data_1",
    icon: "📋",
    is_external: false,
    created_by: 1,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    fields: [],
    field_count: 0,
  };

  const mockFields: Field[] = [
    {
      id: 1,
      app_id: 1,
      field_code: "title",
      field_name: "Title",
      field_type: "text" as FieldType,
      required: true,
      display_order: 2,
      options: {},
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-01-01T00:00:00Z",
    },
    {
      id: 2,
      app_id: 1,
      field_code: "description",
      field_name: "Description",
      field_type: "textarea" as FieldType,
      required: false,
      display_order: 1,
      options: {},
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-01-01T00:00:00Z",
    },
  ];

  beforeEach(() => {
    // Reset store state before each test
    useAppStore.setState({
      currentApp: null,
      currentFields: [],
    });
  });

  describe("setCurrentApp", () => {
    it("should set the current app", () => {
      useAppStore.getState().setCurrentApp(mockApp);

      expect(useAppStore.getState().currentApp).toEqual(mockApp);
    });

    it("should clear the current app when null", () => {
      useAppStore.getState().setCurrentApp(mockApp);
      useAppStore.getState().setCurrentApp(null);

      expect(useAppStore.getState().currentApp).toBeNull();
    });
  });

  describe("setCurrentFields", () => {
    it("should set and sort fields by display_order", () => {
      useAppStore.getState().setCurrentFields(mockFields);

      const fields = useAppStore.getState().currentFields;
      expect(fields).toHaveLength(2);
      // Should be sorted by display_order
      expect(fields[0].field_code).toBe("description"); // display_order: 1
      expect(fields[1].field_code).toBe("title"); // display_order: 2
    });
  });

  describe("updateField", () => {
    it("should update an existing field", () => {
      // Use fresh field data for this test
      const testFields: Field[] = [
        {
          id: 10,
          app_id: 1,
          field_code: "field1",
          field_name: "Field One",
          field_type: "text" as FieldType,
          required: true,
          display_order: 1,
          options: {},
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
      ];

      useAppStore.getState().setCurrentFields(testFields);

      const updatedField = {
        ...testFields[0],
        field_name: "Updated Field One",
      };

      useAppStore.getState().updateField(updatedField);

      const fields = useAppStore.getState().currentFields;
      const field = fields.find((f) => f.id === 10);
      expect(field?.field_name).toBe("Updated Field One");
    });

    it("should not modify other fields", () => {
      // Use fresh field data for this test
      const testFields: Field[] = [
        {
          id: 20,
          app_id: 1,
          field_code: "field_a",
          field_name: "Field A",
          field_type: "text" as FieldType,
          required: true,
          display_order: 1,
          options: {},
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
        {
          id: 21,
          app_id: 1,
          field_code: "field_b",
          field_name: "Field B",
          field_type: "text" as FieldType,
          required: false,
          display_order: 2,
          options: {},
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
      ];

      useAppStore.getState().setCurrentFields(testFields);

      const updatedField = {
        ...testFields[0],
        field_name: "Updated Field A",
      };

      useAppStore.getState().updateField(updatedField);

      const fields = useAppStore.getState().currentFields;
      const otherField = fields.find((f) => f.id === 21);
      expect(otherField?.field_name).toBe("Field B");
    });
  });

  describe("addField", () => {
    it("should add a new field and sort by display_order", () => {
      useAppStore.getState().setCurrentFields(mockFields);

      const newField: Field = {
        id: 3,
        app_id: 1,
        field_code: "status",
        field_name: "Status",
        field_type: "dropdown" as FieldType,
        required: false,
        display_order: 0, // Should be first after sorting
        options: { items: ["Active", "Inactive"] },
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      useAppStore.getState().addField(newField);

      const fields = useAppStore.getState().currentFields;
      expect(fields).toHaveLength(3);
      expect(fields[0].field_code).toBe("status"); // display_order: 0
    });
  });

  describe("removeField", () => {
    it("should remove a field by id", () => {
      useAppStore.getState().setCurrentFields(mockFields);

      useAppStore.getState().removeField(1);

      const fields = useAppStore.getState().currentFields;
      expect(fields).toHaveLength(1);
      expect(fields[0].id).toBe(2);
    });

    it("should not modify store if field does not exist", () => {
      useAppStore.getState().setCurrentFields(mockFields);

      useAppStore.getState().removeField(999);

      const fields = useAppStore.getState().currentFields;
      expect(fields).toHaveLength(2);
    });
  });

  describe("reorderFields", () => {
    it("should reorder fields directly", () => {
      useAppStore.getState().setCurrentFields(mockFields);

      // Reverse the order
      const reorderedFields = [...mockFields].reverse();
      useAppStore.getState().reorderFields(reorderedFields);

      const fields = useAppStore.getState().currentFields;
      expect(fields[0].id).toBe(mockFields[1].id);
      expect(fields[1].id).toBe(mockFields[0].id);
    });
  });
});
