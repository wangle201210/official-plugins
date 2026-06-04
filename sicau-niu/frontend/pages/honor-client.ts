import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function sicauNiuApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface HonorItem {
  id: number;
  honorType: string;
  code: string;
  name: string;
  unlockType: string;
  threshold: number;
  category: string;
  imagePath: string;
  sort: number;
  createdAt: number | null;
  updatedAt: number | null;
}

export interface HonorListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  honorType?: string;
  unlockType?: string;
}

export interface HonorSaveInput {
  honorType: string;
  code: string;
  name: string;
  unlockType: string;
  threshold: number;
  category: string;
  imagePath: string;
  sort: number;
}

export async function listHonors(params?: HonorListParams) {
  const res = await requestClient.get<{
    list: HonorItem[];
    total: number;
  }>(sicauNiuApi("plugins/sicau-niu/admin/honors"), { params });
  return {
    items: res.list,
    total: res.total,
  };
}

export async function getHonor(id: number) {
  return requestClient.get<HonorItem>(
    sicauNiuApi(`plugins/sicau-niu/admin/honors/${id}`),
  );
}

export async function createHonor(values: HonorSaveInput) {
  return requestClient.post<{ id: number }>(
    sicauNiuApi("plugins/sicau-niu/admin/honors"),
    values,
  );
}

export async function updateHonor(id: number, values: HonorSaveInput) {
  return requestClient.put(
    sicauNiuApi(`plugins/sicau-niu/admin/honors/${id}`),
    values,
  );
}

export async function deleteHonor(id: number) {
  return requestClient.delete(
    sicauNiuApi(`plugins/sicau-niu/admin/honors/${id}`),
  );
}
