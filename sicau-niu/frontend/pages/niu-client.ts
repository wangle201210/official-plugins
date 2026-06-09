import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function sicauNiuApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface NiuItem {
  id: number;
  code: string;
  niuType: string;
  specialSubtype: string;
  name: string;
  collegeId: number;
  collegeName: string;
  lat: number;
  lng: number;
  onlineAt: number | null;
  visibleWeekdays: string;
  visibleStart: string;
  visibleEnd: string;
  status: string;
  hasCard: boolean;
  createdAt: number | null;
  updatedAt: number | null;
}

export interface NiuListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  niuType?: string;
}

export interface NiuSaveInput {
  code: string;
  niuType: string;
  specialSubtype: string;
  name: string;
  collegeId: number;
  lat: number;
  lng: number;
  onlineAt: number | null;
  visibleWeekdays: string;
  visibleStart: string;
  visibleEnd: string;
}

export async function listNiu(params?: NiuListParams) {
  const res = await requestClient.get<{
    list: NiuItem[];
    total: number;
  }>(sicauNiuApi("plugins/sicau-niu/admin/niu"), { params });
  return {
    items: res.list,
    total: res.total,
  };
}

export async function getNiu(id: number) {
  return requestClient.get<NiuItem>(
    sicauNiuApi(`plugins/sicau-niu/admin/niu/${id}`),
  );
}

export async function createNiu(values: NiuSaveInput) {
  return requestClient.post<{ id: number }>(
    sicauNiuApi("plugins/sicau-niu/admin/niu"),
    values,
  );
}

export async function updateNiu(id: number, values: NiuSaveInput) {
  return requestClient.put(
    sicauNiuApi(`plugins/sicau-niu/admin/niu/${id}`),
    values,
  );
}

export async function deleteNiu(id: number) {
  return requestClient.delete(
    sicauNiuApi(`plugins/sicau-niu/admin/niu/${id}`),
  );
}
