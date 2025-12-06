import { http, HttpResponse } from "msw";

const API_BASE = "/api/v1";

// Mock data
export const mockUser = {
  id: 1,
  email: "test@example.com",
  name: "Test User",
  role: "admin",
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

export const mockApp = {
  id: 1,
  name: "Test App",
  description: "A test application",
  table_name: "app_data_1",
  icon: "default",
  created_by: 1,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
  fields: [],
  field_count: 0,
};

export const mockField = {
  id: 1,
  app_id: 1,
  field_code: "title",
  field_name: "Title",
  field_type: "text",
  required: true,
  display_order: 1,
  options: {},
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

export const mockRecord = {
  id: 1,
  data: { title: "Test Record" },
  created_by: 1,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

export const mockView = {
  id: 1,
  app_id: 1,
  name: "Default View",
  view_type: "table",
  is_default: true,
  filter_conditions: null,
  sort_conditions: null,
  visible_fields: null,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

export const handlers = [
  // Auth handlers
  http.post(`${API_BASE}/auth/register`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      user: { ...mockUser, email: body.email as string },
      token: "mock-jwt-token",
    });
  }),

  http.post(`${API_BASE}/auth/login`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    if (body.email === "test@example.com" && body.password === "password123") {
      return HttpResponse.json({
        user: mockUser,
        token: "mock-jwt-token",
      });
    }
    return HttpResponse.json({ error: "Invalid credentials" }, { status: 401 });
  }),

  http.get(`${API_BASE}/auth/me`, () => {
    return HttpResponse.json(mockUser);
  }),

  http.post(`${API_BASE}/auth/refresh`, () => {
    return HttpResponse.json({ token: "new-mock-jwt-token" });
  }),

  // Apps handlers
  http.get(`${API_BASE}/apps`, ({ request }) => {
    const url = new URL(request.url);
    const page = parseInt(url.searchParams.get("page") || "1");
    const limit = parseInt(url.searchParams.get("limit") || "20");
    return HttpResponse.json({
      apps: [mockApp],
      pagination: {
        total: 1,
        page,
        limit,
        total_pages: 1,
      },
    });
  }),

  http.get(`${API_BASE}/apps/:id`, ({ params }) => {
    const { id } = params;
    if (id === "999") {
      return HttpResponse.json({ error: "App not found" }, { status: 404 });
    }
    return HttpResponse.json({ ...mockApp, id: parseInt(id as string) });
  }),

  http.post(`${API_BASE}/apps`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockApp,
      name: body.name as string,
      description: body.description as string,
    });
  }),

  http.put(`${API_BASE}/apps/:id`, async ({ request, params }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockApp,
      id: parseInt(params.id as string),
      name: body.name as string,
      description: body.description as string,
    });
  }),

  http.delete(`${API_BASE}/apps/:id`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  // Fields handlers
  http.get(`${API_BASE}/apps/:appId/fields`, () => {
    return HttpResponse.json({ fields: [mockField] });
  }),

  http.post(`${API_BASE}/apps/:appId/fields`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockField,
      field_code: body.field_code as string,
      field_name: body.field_name as string,
      field_type: body.field_type as string,
    });
  }),

  http.put(`${API_BASE}/apps/:appId/fields/:fieldId`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockField,
      field_name: body.field_name as string,
    });
  }),

  http.delete(`${API_BASE}/apps/:appId/fields/:fieldId`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.put(`${API_BASE}/apps/:appId/fields/order`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  // Records handlers
  http.get(`${API_BASE}/apps/:appId/records`, () => {
    return HttpResponse.json({
      records: [mockRecord],
      pagination: {
        total: 1,
        page: 1,
        limit: 20,
        total_pages: 1,
      },
    });
  }),

  http.get(`${API_BASE}/apps/:appId/records/:recordId`, ({ params }) => {
    if (params.recordId === "999") {
      return HttpResponse.json({ error: "Record not found" }, { status: 404 });
    }
    return HttpResponse.json({
      ...mockRecord,
      id: parseInt(params.recordId as string),
    });
  }),

  http.post(`${API_BASE}/apps/:appId/records`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockRecord,
      data: body.data as Record<string, unknown>,
    });
  }),

  http.put(`${API_BASE}/apps/:appId/records/:recordId`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockRecord,
      data: body.data as Record<string, unknown>,
    });
  }),

  http.delete(`${API_BASE}/apps/:appId/records/:recordId`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.post(`${API_BASE}/apps/:appId/records/bulk`, async ({ request }) => {
    const body = (await request.json()) as { records: unknown[] };
    return HttpResponse.json({
      records: body.records.map((_, index) => ({
        ...mockRecord,
        id: index + 1,
      })),
    });
  }),

  http.delete(`${API_BASE}/apps/:appId/records/bulk`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  // Views handlers
  http.get(`${API_BASE}/apps/:appId/views`, () => {
    return HttpResponse.json({ views: [mockView] });
  }),

  http.post(`${API_BASE}/apps/:appId/views`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockView,
      name: body.name as string,
      view_type: body.view_type as string,
    });
  }),

  http.put(`${API_BASE}/apps/:appId/views/:viewId`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockView,
      name: body.name as string,
    });
  }),

  http.delete(`${API_BASE}/apps/:appId/views/:viewId`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  // Charts handlers
  http.post(`${API_BASE}/apps/:appId/charts/data`, () => {
    return HttpResponse.json({
      data: [
        { label: "Category A", value: 10 },
        { label: "Category B", value: 20 },
        { label: "Category C", value: 30 },
      ],
    });
  }),

  http.get(`${API_BASE}/apps/:appId/charts/config`, () => {
    return HttpResponse.json({ configs: [] });
  }),

  http.post(`${API_BASE}/apps/:appId/charts/config`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      id: 1,
      app_id: 1,
      name: body.name as string,
      chart_type: body.chart_type as string,
      config: body,
      created_by: 1,
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-01-01T00:00:00Z",
    });
  }),

  http.delete(`${API_BASE}/apps/:appId/charts/config/:configId`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  // Profile handlers
  http.put(`${API_BASE}/auth/profile`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockUser,
      name: body.name as string,
    });
  }),

  http.put(`${API_BASE}/auth/password`, () => {
    return HttpResponse.json({ message: "パスワードを変更しました" });
  }),

  // Users handlers (admin only)
  http.get(`${API_BASE}/users`, ({ request }) => {
    const url = new URL(request.url);
    const page = parseInt(url.searchParams.get("page") || "1");
    const limit = parseInt(url.searchParams.get("limit") || "20");
    return HttpResponse.json({
      users: [mockUser],
      pagination: {
        total: 1,
        page,
        limit,
        total_pages: 1,
      },
    });
  }),

  http.get(`${API_BASE}/users/:id`, ({ params }) => {
    if (params.id === "999") {
      return HttpResponse.json({ error: "User not found" }, { status: 404 });
    }
    return HttpResponse.json({
      ...mockUser,
      id: parseInt(params.id as string),
    });
  }),

  http.post(`${API_BASE}/users`, async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockUser,
      id: 2,
      email: body.email as string,
      name: body.name as string,
      role: body.role as string,
    });
  }),

  http.put(`${API_BASE}/users/:id`, async ({ request, params }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json({
      ...mockUser,
      id: parseInt(params.id as string),
      name: body.name as string,
      role: body.role as string,
    });
  }),

  http.delete(`${API_BASE}/users/:id`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  // Dashboard handlers
  http.get(`${API_BASE}/dashboard/stats`, () => {
    return HttpResponse.json({
      stats: {
        app_count: 5,
        total_records: 100,
        user_count: 10,
        todays_updates: 3,
      },
    });
  }),
];
