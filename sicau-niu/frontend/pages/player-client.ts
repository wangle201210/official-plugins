import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "sicau-niu";

function sicauNiuApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface PlayerItem {
  id: number;
  nickname: string;
  phone: string;
  identityType: string;
  collegeId: number;
  grade: number;
  graduationYear: number;
  createdAt: number | null;
  updatedAt: number | null;
}

export interface PlayerListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  identityType?: string;
}

export async function listPlayers(params?: PlayerListParams) {
  const res = await requestClient.get<{
    list: PlayerItem[];
    total: number;
  }>(sicauNiuApi("plugins/sicau-niu/admin/players"), { params });
  return {
    items: res.list,
    total: res.total,
  };
}
