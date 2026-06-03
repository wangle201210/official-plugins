import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function sicauNiuApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface QuoteItem {
  id: number;
  content: string;
  enabled: number;
  createdAt: number | null;
  updatedAt: number | null;
}

export interface QuoteListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
}

export interface QuoteSaveInput {
  content: string;
  enabled: number;
}

export async function listQuotes(params?: QuoteListParams) {
  const res = await requestClient.get<{
    list: QuoteItem[];
    total: number;
  }>(sicauNiuApi("plugins/sicau-niu/admin/quotes"), { params });
  return {
    items: res.list,
    total: res.total,
  };
}

export async function createQuote(values: QuoteSaveInput) {
  return requestClient.post<{ id: number }>(
    sicauNiuApi("plugins/sicau-niu/admin/quotes"),
    values,
  );
}

export async function updateQuote(id: number, values: QuoteSaveInput) {
  return requestClient.put(
    sicauNiuApi(`plugins/sicau-niu/admin/quotes/${id}`),
    values,
  );
}

export async function deleteQuote(id: number) {
  return requestClient.delete(
    sicauNiuApi(`plugins/sicau-niu/admin/quotes/${id}`),
  );
}
