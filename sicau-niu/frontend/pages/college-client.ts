import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function sicauNiuApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface CollegeItem {
  id: number;
  name: string;
  sort: number;
  createdAt: number | null;
  updatedAt: number | null;
}

export interface CollegeListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
}

export interface CollegeSaveInput {
  name: string;
  sort: number;
}

export async function listColleges(params?: CollegeListParams) {
  const res = await requestClient.get<{
    list: CollegeItem[];
    total: number;
  }>(sicauNiuApi("plugins/sicau-niu/admin/colleges"), { params });
  return {
    items: res.list,
    total: res.total,
  };
}

export async function createCollege(values: CollegeSaveInput) {
  return requestClient.post<{ id: number }>(
    sicauNiuApi("plugins/sicau-niu/admin/colleges"),
    values,
  );
}

export async function updateCollege(id: number, values: CollegeSaveInput) {
  return requestClient.put(
    sicauNiuApi(`plugins/sicau-niu/admin/colleges/${id}`),
    values,
  );
}

export async function deleteCollege(id: number) {
  return requestClient.delete(
    sicauNiuApi(`plugins/sicau-niu/admin/colleges/${id}`),
  );
}
