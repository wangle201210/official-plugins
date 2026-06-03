import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function sicauNiuApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface IronItem {
  id: number;
  code: string;
  name: string;
  lastLat: number;
  lastLng: number;
  locatedAt: number | null;
  remark: string;
  createdAt: number | null;
  updatedAt: number | null;
}

export interface IronListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
}

export interface IronSaveInput {
  code: string;
  name: string;
  remark: string;
}

export async function listIron(params?: IronListParams) {
  const res = await requestClient.get<{
    list: IronItem[];
    total: number;
  }>(sicauNiuApi("plugins/sicau-niu/admin/iron"), { params });
  return {
    items: res.list,
    total: res.total,
  };
}

export async function createIron(values: IronSaveInput) {
  return requestClient.post<{ id: number }>(
    sicauNiuApi("plugins/sicau-niu/admin/iron"),
    values,
  );
}

export async function updateIron(id: number, values: IronSaveInput) {
  return requestClient.put(
    sicauNiuApi(`plugins/sicau-niu/admin/iron/${id}`),
    values,
  );
}

export async function deleteIron(id: number) {
  return requestClient.delete(
    sicauNiuApi(`plugins/sicau-niu/admin/iron/${id}`),
  );
}
