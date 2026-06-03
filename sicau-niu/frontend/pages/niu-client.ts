import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function sicauNiuApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface CattleItem {
  id: number;
  name: string;
  breed: string;
  weightKg: number;
  createdAt: number | null;
}

export interface CattleListParams {
  keyword?: string;
}

export async function listCattle(params?: CattleListParams) {
  const res = await requestClient.get<{
    list: CattleItem[];
    total: number;
  }>(sicauNiuApi("plugins/sicau-niu/cattle"), { params });
  return {
    items: res.list,
    total: res.total,
  };
}
